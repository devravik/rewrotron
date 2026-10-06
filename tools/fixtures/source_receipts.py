#!/usr/bin/env python3
"""Independent local receipt/ACK transport timing; completion remains unavailable."""
import argparse
import base64
import copy
import hashlib
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
import ipaddress
import json
from pathlib import Path
import shutil
import tempfile
import threading
import time
import uuid

import ci3_reference as fixture
import source_causal as causal
import source_roles as roles


class Receiver:
    """Supervisor-owned receiver: producer has HTTP delivery only, no evidence port."""
    def __init__(self, address, storage, scope, initiating, worker_span, anchor):
        self.storage=storage;storage.mkdir(mode=0o700)
        self.scope=copy.deepcopy(scope);self.initiating=copy.deepcopy(initiating)
        self.worker_span=worker_span;self.anchor=copy.deepcopy(anchor)
        self.events=[];self.lock=threading.Lock();self.queued=None
        receiver=self
        class Handler(BaseHTTPRequestHandler):
            def log_message(self,*args):
                pass
            def do_POST(self):
                try:
                    size=int(self.headers.get('Content-Length','0'))
                    if size<1 or size>65536:raise ValueError('body bound')
                    body=self.rfile.read(size)
                    payload=json.loads(body)
                    envelope=json.loads(base64.b64decode(self.headers.get('X-Rewrotron-Capture',''),validate=True))
                    if set(payload)!={'business_key','item'} or payload['business_key']!='item-'+str(payload['item']):raise ValueError('invalid actual payload')
                    if receiver.queued is None:raise ValueError('queue binding unavailable')
                    causal.validate_chain(receiver.initiating,receiver.queued,envelope,receiver.scope['execution_id'],receiver.worker_span)
                    if envelope['job_id']!='job_'+str(receiver.job_id) or payload['item']!=receiver.item_id:raise ValueError('wrong actual queued job/payload')
                    stamp=time.monotonic_ns()
                    if stamp<receiver.anchor['monotonic_ns']:raise ValueError('wrong timing anchor')
                    business_id=hashlib.sha256(payload['business_key'].encode()).hexdigest()
                    with receiver.lock:
                        target=receiver.storage/(business_id+'.json')
                        if target.exists():raise ValueError('replayed delivery not qualified by this profile')
                        receipt_id='receipt_'+uuid.uuid4().hex;ack_id='ack_'+uuid.uuid4().hex
                        accepted={**receiver.scope,'receipt_id':receipt_id,'business_application_id':business_id,
                                  'payload':payload,'capture_context':envelope,'received_monotonic_ns':stamp,
                                  'elapsed_ms':(stamp-receiver.anchor['monotonic_ns']+999999)//1000000}
                        with target.open('xb') as output:
                            output.write(fixture.canonical(accepted));output.flush()
                            __import__('os').fsync(output.fileno())
                        durable=time.monotonic_ns()
                        receiver.events.append({**accepted,'phase':'RECEIPT_DURABLY_ACCEPTED',
                            'accepted_monotonic_ns':durable,'elapsed_ms':(durable-receiver.anchor['monotonic_ns']+999999)//1000000,
                            'durable_artifact_ref':fixture.ref(target.read_bytes())})
                    response=fixture.canonical({'receipt':receipt_id,'ack_id':ack_id,'business_application_id':business_id,'applied':True})
                    self.send_response(200);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(response)));self.end_headers()
                    self.wfile.write(response);self.wfile.flush()
                    stamp=time.monotonic_ns()
                    with receiver.lock:
                        receiver.events.append({**receiver.scope,'phase':'ACK_RESPONSE_SENT','receipt_id':receipt_id,
                            'ack_id':ack_id,'business_application_id':business_id,'response_base64':base64.b64encode(response).decode(),
                            'sent_monotonic_ns':stamp,'elapsed_ms':(stamp-receiver.anchor['monotonic_ns']+999999)//1000000})
                except (ValueError,KeyError,TypeError):
                    self.send_error(409)
        self.server=ThreadingHTTPServer((address,0),Handler)
        self.thread=threading.Thread(target=self.server.serve_forever,daemon=True)
    def bind_queue(self,row):
        self.queued=json.loads(row['capture_context']);self.job_id=row['id'];self.item_id=row['item']
    def __enter__(self):
        self.thread.start();return self
    def __exit__(self,*args):
        self.server.shutdown();self.server.server_close();self.thread.join()


def validate_events(events,scope,anchor,expected_ref):
    if fixture.ref(fixture.canonical(events))!=expected_ref:raise ValueError('receiver evidence altered')
    if len(events)!=2 or [e['phase'] for e in events]!=['RECEIPT_DURABLY_ACCEPTED','ACK_RESPONSE_SENT']:
        raise ValueError('missing/replayed receipt or ACK')
    received,sent=events
    for event in events:
        if any(event.get(k)!=v for k,v in scope.items()):raise ValueError('wrong execution/window/source/runtime/build')
        stamp=event.get('accepted_monotonic_ns',event.get('sent_monotonic_ns'))
        if stamp<anchor['monotonic_ns'] or event['elapsed_ms']!=(stamp-anchor['monotonic_ns']+999999)//1000000:
            raise ValueError('wrong dispatch-anchor timing')
    if received['accepted_monotonic_ns']<received['received_monotonic_ns'] or sent['sent_monotonic_ns']<received['accepted_monotonic_ns'] or sent['receipt_id']!=received['receipt_id'] or sent['business_application_id']!=received['business_application_id']:
        raise ValueError('ACK references wrong receipt')
    response=json.loads(base64.b64decode(sent['response_base64']))
    if response['receipt']!=sent['receipt_id'] or response['ack_id']!=sent['ack_id'] or response['business_application_id']!=sent['business_application_id']:
        raise ValueError('altered ACK body')


def run(out):
    out.mkdir(mode=0o700)
    impl_ref=fixture.store(out,'implementation.json',{'source_receipts.py':fixture.ref(Path(__file__).read_bytes()),
        'source_causal.py':fixture.ref(Path(causal.__file__).read_bytes()),'source_roles.py':fixture.ref(Path(roles.__file__).read_bytes()),
        'ci3_reference.py':fixture.ref(Path(fixture.__file__).read_bytes())})
    suite_ref=fixture.store(out,'control-suite.json',{'implementation_ref':impl_ref,'required':['independent-durable-receipt','actual-wire-ack-response',
        'dispatch-anchor-binding','withheld-worker-no-receipt','producer-storage-write-denied','receiver-forged-envelope-refused'],
        'limits':'ACK send is not ACK consumption or completion; no horizon/quiescence'})
    reports,builds=[],[]
    with tempfile.TemporaryDirectory(prefix='rwt-receipt-source-') as temporary:
        root=Path(temporary)/'source';original,lock=causal.seal(root)
        worker=root/'worker.php'
        worker.write_text(worker.read_text().replace("'http://sink:8081/'","getenv('RWT_RECEIVER_URL')"))
        lock['source_files']['worker.php']=fixture.ref(worker.read_bytes())[7:]
        (root/'runtime.lock.json').write_bytes(fixture.canonical(lock)+b'\n');fixture.pins(root)
        fixture.store(out,'runtime.lock.json',lock)
        fixture.store(out,'worker.php.json',{'path':'worker.php','bytes_base64':base64.b64encode(worker.read_bytes()).decode()})
        original_ref=fixture.store(out,'original-source-manifest.json',original)
        source_ref=fixture.store(out,'source-manifest.json',fixture.source_manifest(root))
        runtime_ref=fixture.ref((root/'runtime.lock.json').read_bytes())
        profile_ref=fixture.store(out,'input-profile.json',{'side':'SOURCE','source_manifest_ref':source_ref,'runtime_ref':runtime_ref,
            'implementation_ref':impl_ref,'boundary':'independent receiver transport receipt/send timing only','horizon':None,'quiescence':None})
        for index in range(2):
            image,build_ref,runtime=fixture.build(lock,'rewrotron-ci3:receipts-'+str(index),out,root);builds.append(build_ref)
            execution='source_receipts_'+uuid.uuid4().hex;run_dir=out/execution;run_dir.mkdir()
            scope={'side':'SOURCE','execution_id':execution,'window_id':'window_'+uuid.uuid4().hex,
                'source_manifest_ref':source_ref,'runtime_ref':runtime_ref,'build_ref':build_ref}
            with fixture.SourceRuntime(image,lock,execution,root) as app:
                network=json.loads(fixture.docker('network','inspect',app.network).stdout)[0]
                gateway=network['IPAM']['Config'][0]['Gateway'];subnet=ipaddress.ip_network(network['IPAM']['Config'][0]['Subnet'])
                if ipaddress.ip_address(gateway) not in subnet or not network['Internal']:raise ValueError('receiver outside owned internal namespace')
                source=roles.SourceCapture(fixture.Observer(app).snapshot,source_ref)
                _,role_controls=roles.controls(source,source_ref)
                driver=fixture.Driver(app);login=driver.request('/login',{'name':'alice','password':'fixture-password'})
                cookie=fixture.cookie(login);token=fixture.body(driver.request('/csrf',cookie=cookie))['token']
                child='worker_'+uuid.uuid4().hex
                initiating={'project_id':'ci3_reference','run_id':execution,'workflow_instance_id':execution,'operation_id':'create',
                    'request_id':'request_'+uuid.uuid4().hex,'process_id':str(app.launch['State']['Pid']),'trace_id':'trace_'+uuid.uuid4().hex,
                    'span_id':'create_'+uuid.uuid4().hex,'parent_span_id':None,'cause_ids':[],'job_id':None,'effect_id':None}
                anchor={'monotonic_ns':time.monotonic_ns(),'utc_ns':time.time_ns()}
                storage=run_dir/'receiver-store'
                with Receiver(gateway,storage,scope,initiating,child,anchor) as receiver:
                    # Anchor immediately at initiating dispatch, after receiver startup.
                    anchor={'monotonic_ns':time.monotonic_ns(),'utc_ns':time.time_ns()}
                    receiver.anchor=copy.deepcopy(anchor)
                    response=driver.request('/create',{'name':'receipt-control','csrf':token,'_rwt_capture':fixture.canonical(initiating).decode()},cookie)
                    queued=causal.read_jobs(app)
                    if response['status']!=202 or queued[0]['status']!='pending' or receiver.events:raise RuntimeError('withheld worker control failed')
                    receiver.bind_queue(queued[0])
                    # Producer has no host mounts/storage credentials or event-writing API.
                    producer=json.loads(fixture.docker('inspect',app.app).stdout)[0]
                    if producer['Mounts'] and any(m['Type']=='bind' for m in producer['Mounts']):raise RuntimeError('producer mounts receiver evidence')
                    fake=fixture.docker('exec',app.app,'php','-r',"exit(@file_put_contents('/supervisor-receiver-evidence','ACK=true')===false?0:1);",check=False)
                    if fake.returncode!=0 or receiver.events:raise RuntimeError('producer fabricated receiver evidence')
                    url='http://'+gateway+':'+str(receiver.server.server_port)+'/'
                    fraud_script = "$p=json_decode(stream_get_contents(STDIN),true);$c=curl_init($p['url']);curl_setopt($c,CURLOPT_POST,true);curl_setopt($c,CURLOPT_RETURNTRANSFER,true);curl_setopt($c,CURLOPT_POSTFIELDS,'{\"business_key\":\"item-1\",\"item\":1}');curl_setopt($c,CURLOPT_HTTPHEADER,['Content-Type: application/json','X-Rewrotron-Capture: '.base64_encode('{\"ACK\":true}')]);curl_exec($c);echo curl_getinfo($c,CURLINFO_HTTP_CODE);"
                    refused=fixture.docker('exec','-i',app.app,'php','-r',fraud_script,data=fixture.canonical({'url':url}))
                    if refused.stdout!=b'409' or receiver.events:raise RuntimeError('producer fabricated ACK accepted')
                    result=fixture.docker('exec','--env','RWT_RECEIVER_URL='+url,'--env','RWT_WORKER_SPAN='+child,app.app,'php','/app/worker.php',check=False)
                    events=copy.deepcopy(receiver.events)
                    if result.returncode!=0:raise RuntimeError('worker receiver transport failed')
                    events_ref=fixture.store(run_dir,'receiver-events.json',events)
                    validate_events(events,scope,anchor,events_ref)
                    final=causal.read_jobs(app)
                    if final[0]['receipt']!=events[0]['receipt_id'] or final[0]['status']!='acked':raise RuntimeError('worker stored receipt mismatch')
                    for key in ('execution_id','window_id','runtime_ref','source_manifest_ref','received_monotonic_ns'):
                        bad=copy.deepcopy(events);bad[0][key]='foreign'
                        try:validate_events(bad,scope,anchor,events_ref)
                        except ValueError:continue
                        raise RuntimeError('altered receiver evidence accepted')
                    report={**scope,'original_source_ref':original_ref,'anchor':anchor,'initiating':initiating,'worker_span':child,
                        'queued_rows':queued,'final_rows':final,'receiver_events_ref':events_ref,'role_controls':role_controls,
                        'producer_forged_ack_http_status':409,'receiver_endpoint':url,'receiver_process_id':str(__import__('os').getpid()),
                        'launch_ref':fixture.store(run_dir,'launch.json',app.launch),'runtime':runtime,'implementation_ref':impl_ref,
                        'receipt':'INDEPENDENT_DURABLE_ACCEPTANCE_CAPTURED','ack':'RESPONSE_SEND_CAPTURED_CONSUMPTION_UNQUALIFIED',
                        'completion':'UNKNOWN','horizon':None,'quiescence':None,'due_beyond_horizon':None,
                        'async_causal_completion_abi':'PARTIAL','db_intermediate_visibility':'PARTIAL'}
                    fixture.store(run_dir,'capture-report.json',report);reports.append(report)
        if builds[0]!=builds[1]:raise RuntimeError('independent build mismatch')
        result_ref=fixture.store(out,'control-result.json',{'runs':reports,'reproduction_passed':True})
        q={'schema_version':'1.0','qualification_id':'ci3_receipt_transport_timing_1','project_id':'ci3_reference','subject_kind':'ENGINE',
            'implementation_ref':impl_ref,'schema_refs':[fixture.ref((fixture.REPO/'specs/observation/effect-occurrence.schema.json').read_bytes())],
            'capability_ids':['source-independent-receipt-ack-send-timing'],'input_profile_ref':profile_ref,'control_suite_ref':suite_ref,
            'control_result_refs':[result_ref],'status':'SUPPORTED','reason_codes':['EXECUTED_INDEPENDENT_RECEIVER_BOUNDARIES'],
            'limitations':['Transport acceptance/send only; no independently witnessed worker ACK consumption',
                'No qualified horizon/due/quiescence/completion; aggregate async ABI and DB intermediate visibility PARTIAL',
                'New disposable source revision only; original B1 and earlier causal revision not upgraded']}
        print(fixture.store(out,'qualification.json',q),flush=True)


if __name__=='__main__':
    parser=argparse.ArgumentParser(description=__doc__);parser.add_argument('output',type=Path);run(parser.parse_args().output)
