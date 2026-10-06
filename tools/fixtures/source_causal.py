#!/usr/bin/env python3
"""Qualify SOURCE job-envelope propagation in a separate disposable revision."""
import argparse
import copy
import json
from pathlib import Path
import shutil
import tempfile
import uuid

import ci3_reference as fixture
import source_roles as roles


def seal(root):
    shutil.copytree(fixture.FIXTURE, root, ignore=shutil.ignore_patterns('README.md', 'evidence'))
    fixture.pins(root)
    original = fixture.source_manifest(root)
    overlay = Path(__file__).with_name('causal-overlay')
    changes = {'seed.sql':'seed.sql', 'Reference.php':'application/controllers/Reference.php',
               'worker.php':'worker.php', 'sink.php':'sink.php'}
    lock = json.loads((root/'runtime.lock.json').read_bytes())
    for source, target in changes.items():
        raw = (overlay/source).read_bytes(); (root/target).write_bytes(raw)
        lock['source_files'][target] = fixture.ref(raw)[7:]
    (root/'runtime.lock.json').write_bytes(fixture.canonical(lock)+b'\n')
    fixture.pins(root)
    return original, lock


def validate_chain(initiating, queued, received, expected_run, expected_span):
    allowed = {'project_id','run_id','workflow_instance_id','operation_id','request_id','process_id',
               'trace_id','span_id','parent_span_id','cause_ids','job_id','effect_id'}
    if any(set(e) != allowed for e in (initiating,queued,received)):
        raise ValueError('missing/foreign causal envelope fields')
    if any(e['run_id'] != expected_run or e['project_id'] != 'ci3_reference' for e in (initiating,queued,received)):
        raise ValueError('foreign execution/project')
    for key in ('workflow_instance_id','operation_id','request_id','trace_id'):
        if initiating[key] != queued[key] or queued[key] != received[key]:
            raise ValueError('wrong initiating relationship')
    if queued['span_id'] != initiating['span_id'] or queued['parent_span_id'] != initiating['parent_span_id'] or queued['cause_ids'] != initiating['cause_ids']:
        raise ValueError('queue altered initiating cause')
    if not queued['job_id'] or not queued['effect_id'] or any(received[k] != queued[k] for k in ('job_id','effect_id')):
        raise ValueError('wrong/missing job or effect')
    if received['parent_span_id'] != initiating['span_id'] or received['cause_ids'] != [initiating['span_id']] or received['span_id'] != expected_span or not received['process_id'].isdigit():
        raise ValueError('missing/wrong worker propagation')


def read_jobs(app):
    query = "SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT('id',id,'item',item,'status',status,'receipt',receipt,'capture_context',capture_context)),JSON_ARRAY()) FROM fixture.jobs;"
    return json.loads(fixture.docker('exec',app.db,'mysql','-uobserver','-psynthetic-observer-only',
                                    '--batch','--raw','--skip-column-names','-e',query).stdout)


def run(out):
    out.mkdir(mode=0o700)
    implementation_ref = fixture.store(out,'implementation.json', {
        'source_causal.py':fixture.ref(Path(__file__).read_bytes()),
        'source_roles.py':fixture.ref(Path(roles.__file__).read_bytes()),
        'ci3_reference.py':fixture.ref(Path(fixture.__file__).read_bytes()),
        'overlay':{p.name:fixture.ref(p.read_bytes()) for p in Path(__file__).with_name('causal-overlay').iterdir()}})
    suite_ref = fixture.store(out,'control-suite.json',{'required':['persisted-producer-envelope','worker-to-receiver-carrier',
        'wrong-cause-refused','cross-run-refused','withheld-worker-not-completion','source-role-controls'],
        'implementation_ref':implementation_ref})
    reports, builds = [], []
    with tempfile.TemporaryDirectory(prefix='rwt-causal-revision-') as temporary:
        root = Path(temporary)/'source'
        original, lock = seal(root)
        original_ref = fixture.store(out,'original-source-manifest.json',original)
        source_ref = fixture.store(out,'source-manifest.json',fixture.source_manifest(root))
        runtime_ref = fixture.ref((root/'runtime.lock.json').read_bytes())
        fixture.store(out,'runtime.lock.json',lock)
        profile_ref = fixture.store(out,'input-profile.json',{'side':'SOURCE','fixture':'ci3-causal-disposable-1',
            'original_source_ref':original_ref,'source_manifest_ref':source_ref,'runtime_ref':runtime_ref,
            'scope':'job-envelope propagation only; no receipt timing or completion', 'implementation_ref':implementation_ref})
        for index in range(2):
            image, build_ref, runtime = fixture.build(lock,'rewrotron-ci3:causal-'+str(index),out,root)
            builds.append(build_ref)
            run_id = 'source_causal_'+uuid.uuid4().hex
            run_dir = out/run_id;run_dir.mkdir()
            with fixture.SourceRuntime(image,lock,run_id,root) as app:
                source = roles.SourceCapture(fixture.Observer(app).snapshot,source_ref)
                _, role_controls = roles.controls(source,source_ref)
                driver = fixture.Driver(app)
                login=driver.request('/login',{'name':'alice','password':'fixture-password'})
                cookie=fixture.cookie(login);token=fixture.body(driver.request('/csrf',cookie=cookie))['token']
                parent='source_create_'+uuid.uuid4().hex;child='source_worker_'+uuid.uuid4().hex
                initiating={'project_id':'ci3_reference','run_id':run_id,'workflow_instance_id':run_id,
                    'operation_id':'create','request_id':'request_'+uuid.uuid4().hex,
                    'process_id':str(app.launch['State']['Pid']),'trace_id':'trace_'+uuid.uuid4().hex,
                    'span_id':parent,'parent_span_id':None,'cause_ids':[],'job_id':None,'effect_id':None}
                response=driver.request('/create',{'name':'causal-fixture','csrf':token,'_rwt_capture':fixture.canonical(initiating).decode()},cookie)
                queued_rows=read_jobs(app)
                pending_state=json.loads(source.capture({'side':'SOURCE','source_manifest_ref':source_ref}))
                if response['status']!=202 or len(queued_rows)!=1 or queued_rows[0]['status']!='pending' or pending_state['sink_files'] not in ({},[]):
                    raise RuntimeError('withheld worker control failed')
                queued=json.loads(queued_rows[0]['capture_context'])
                worker=fixture.docker('exec','--env','RWT_WORKER_SPAN='+child,app.app,'php','/app/worker.php',check=False)
                final_rows=read_jobs(app)
                final_state=json.loads(source.capture({'side':'SOURCE','source_manifest_ref':source_ref}))
                if worker.returncode!=0 or len(final_state['sink_files'])!=1 or final_rows[0]['status']!='acked':
                    raise RuntimeError('worker propagation execution failed')
                acceptance=json.loads(next(iter(final_state['sink_files'].values()))['content'])
                received=acceptance['capture_context']
                validate_chain(initiating,queued,received,run_id,child)
                if acceptance['business_application_id']!=final_rows[0]['receipt'] or acceptance['payload']['item']!=queued_rows[0]['item']:
                    raise RuntimeError('sink/job relationship mismatch')
                for key in ('run_id','parent_span_id','cause_ids','job_id','span_id'):
                    altered=copy.deepcopy(received);altered[key]=['foreign'] if key=='cause_ids' else 'foreign'
                    try:validate_chain(initiating,queued,altered,run_id,child)
                    except ValueError:continue
                    raise RuntimeError('foreign propagation accepted')
                report={'side':'SOURCE','run_id':run_id,'source_manifest_ref':source_ref,'original_source_ref':original_ref,
                    'runtime_ref':runtime_ref,'build_ref':build_ref,'implementation_ref':implementation_ref,
                    'initiating':initiating,'queued':queued,'received':received,'worker_span':child,
                    'http_response':response,'queued_rows':queued_rows,'final_rows':final_rows,
                    'pending_state':pending_state,'final_state':final_state,'role_controls':role_controls,
                    'launch_ref':fixture.store(run_dir,'launch.json',app.launch),'runtime':runtime,
                    'propagation':'SUPPORTED_WITHIN_CAUSAL_REVISION','completion':'UNKNOWN',
                    'horizon':None,'due_beyond_horizon':None,'async_causal_completion_abi':'PARTIAL',
                    'db_intermediate_visibility':'PARTIAL'}
                fixture.store(run_dir,'capture-report.json',report);reports.append(report)
        if builds[0]!=builds[1]:raise RuntimeError('independent build mismatch')
        result_ref=fixture.store(out,'control-result.json',{'runs':reports,'reproduction_passed':True})
        q={'schema_version':'1.0','qualification_id':'ci3_causal_carrier_1','project_id':'ci3_reference','subject_kind':'ENGINE',
            'implementation_ref':implementation_ref,'schema_refs':[fixture.ref((fixture.REPO/'specs/observation/causality.schema.json').read_bytes())],
            'capability_ids':['source-job-envelope-propagation'],'input_profile_ref':profile_ref,'control_suite_ref':suite_ref,
            'control_result_refs':[result_ref],'status':'SUPPORTED','reason_codes':['EXECUTED_PRODUCER_QUEUE_WORKER_RECEIVER_CARRIER'],
            'limitations':['New disposable source revision only; original B1 not upgraded','Carrier only; worker process field is application-emitted, not independently witnessed',
                           'No independent wire ACK timing, qualified horizon/quiescence or completion','DB intermediate and async causal/completion ABI remain PARTIAL']}
        print(fixture.store(out,'qualification.json',q),flush=True)


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('output',type=Path);run(parser.parse_args().output)
