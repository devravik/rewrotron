#!/usr/bin/env python3
"""Measure existing CI3 async stages; fail closed on unavailable completion ABI."""
import argparse
import copy
import json
from pathlib import Path
import shutil
import tempfile
import uuid

import ci3_reference as fixture
import source_horizon as horizon
import source_roles as roles
import source_windows as windows


GAPS = ('missing-queue-causal-envelope', 'missing-worker-parent-span',
        'missing-independent-wire-receipt-and-ack', 'missing-initiating-dispatch-delivery-timing',
        'missing-qualified-due-schedule-and-quiescence')


def assess(record, scope, expected_ref):
    """Current raw-stage ABI cannot satisfy X12, even with a stored acked row."""
    windows.validate(record, scope, expected_ref)
    for phase in ('pre_capture', 'post_capture'):
        value = record[phase]
        if set(value) != {'actual_state', 'queue'} or value['actual_state']['side'] != 'SOURCE':
            raise ValueError('missing/foreign actual capture')
        census = value['queue']
        if census['side'] != 'SOURCE' or census['capture_method'] != 'fixed-read-only-server-census':
            raise ValueError('unqualified queue capture')
        if fixture.ref(census['raw'].encode()) != census['raw_ref'] or json.loads(census['raw']) != census['observed']:
            raise ValueError('queue evidence changed')
    # Neither a response nor a queue/sink snapshot is the missing propagation ABI.
    # Unknown fields cannot smuggle self-authored completion into this reader.
    return {'outcome': 'UNKNOWN', 'reason_codes': list(GAPS),
            'observed_queue_before': record['pre_capture']['queue']['observed'],
            'observed_queue_after': record['post_capture']['queue']['observed'],
            'http_or_process_result': record['result'],
            'schedule_horizon_ms': None, 'due_beyond_horizon': None,
            'async_causal_completion_abi': 'PARTIAL', 'db_intermediate_visibility': 'PARTIAL'}


def run(out):
    out.mkdir(mode=0o700)
    role_root = fixture.FIXTURE / 'evidence/roles/run-0'
    role = json.loads((role_root / 'qualification.json').read_bytes())
    role_ref = fixture.ref(fixture.canonical(role))
    role_index = json.loads((role_root / 'artifact-index.json').read_bytes())
    implementation = json.loads((role_root / 'implementation.json').read_bytes())
    if role_index['qualification_ref'] != role_ref or role['status'] != 'SUPPORTED' or implementation != {
            'source_roles.py': fixture.ref(Path(roles.__file__).read_bytes()),
            'ci3_reference.py': fixture.ref(Path(fixture.__file__).read_bytes())}:
        raise ValueError('stale role qualification')
    impl_ref = fixture.store(out, 'implementation.json', {**implementation,
        'source_async.py': fixture.ref(Path(__file__).read_bytes()),
        'source_windows.py': fixture.ref(Path(windows.__file__).read_bytes()),
        'source_horizon.py': fixture.ref(Path(horizon.__file__).read_bytes())})
    suite_ref = fixture.store(out, 'control-suite.json', {'implementation_ref': impl_ref,
        'required': ['http-202-withheld-worker-not-complete', 'worker-exit-and-stored-receipt-not-complete',
                     'altered-capture-refused', 'independent-pinned-build-reproduction'],
        'boundary': 'actual raw queue/worker/sink snapshots; missing causal/completion ABI remains UNKNOWN'})
    reports, builds = [], []
    with tempfile.TemporaryDirectory(prefix='rwt-async-sealed-') as temporary:
        root = Path(temporary) / 'source'
        shutil.copytree(fixture.FIXTURE, root, ignore=shutil.ignore_patterns('evidence', 'README.md'))
        lock = fixture.pins(root)
        source_ref = fixture.store(out, 'source-manifest.json', fixture.source_manifest(root))
        runtime_ref = fixture.ref((root / 'runtime.lock.json').read_bytes())
        profile_ref = fixture.store(out, 'input-profile.json', {'side': 'SOURCE', 'fixture': 'ci3-reference-b1',
            'source_manifest_ref': source_ref, 'runtime_ref': runtime_ref, 'role_isolation_ref': role_ref,
            'implementation_ref': impl_ref, 'scope': 'existing create/pending job/manual worker/local sink',
            'completion': 'UNAVAILABLE', 'schedule_horizon_ms': None, 'due_beyond_horizon': None})
        for index in range(2):
            image, build_ref, runtime = fixture.build(lock, 'rewrotron-ci3:async-' + str(index), out, root)
            builds.append(build_ref)
            execution_id = 'source_async_' + uuid.uuid4().hex
            run_dir = out / execution_id; run_dir.mkdir()
            scope = {'side': 'SOURCE', 'execution_id': execution_id, 'source_manifest_ref': source_ref,
                     'runtime_ref': runtime_ref, 'build_ref': build_ref}
            with fixture.SourceRuntime(image, lock, execution_id, root) as app:
                source = roles.SourceCapture(fixture.Observer(app).snapshot, source_ref)
                def snapshot():
                    return {'actual_state': json.loads(source.capture({'side': 'SOURCE', 'source_manifest_ref': source_ref})),
                            'queue': horizon.queue(app)}
                driver = fixture.Driver(app)
                login = driver.request('/login', {'name': 'alice', 'password': 'fixture-password'})
                cookie = fixture.cookie(login)
                token = fixture.body(driver.request('/csrf', cookie=cookie))['token']
                observer = windows.Windows(scope, snapshot, suite_ref)
                result, pending_id = observer.invoke('HTTP /create', lambda: driver.request('/create', {'name': 'async-control', 'csrf': token}, cookie))
                pending = json.loads(observer.records[pending_id]); pending_ref = fixture.store(run_dir, 'pending.json', pending)
                before_worker = assess(pending, scope, pending_ref)
                jobs = before_worker['observed_queue_after']['jobs']
                if result['status'] != 202 or len(jobs) != 1 or jobs[0]['status'] != 'pending' or jobs[0]['receipt'] is not None or pending['post_capture']['actual_state']['sink_files'] not in ({}, []):
                    raise RuntimeError('HTTP success without delivery control did not execute')
                if before_worker['outcome'] != 'UNKNOWN':
                    raise RuntimeError('HTTP success manufactured async completion')
                exit_code, worker_id = observer.invoke('PROCESS worker', driver.worker)
                worker = json.loads(observer.records[worker_id]); worker_ref = fixture.store(run_dir, 'worker.json', worker)
                after_worker = assess(worker, scope, worker_ref)
                jobs_after = after_worker['observed_queue_after']['jobs']
                sinks = worker['post_capture']['actual_state']['sink_files']
                if exit_code != 0 or len(jobs_after) != 1 or jobs_after[0]['status'] != 'acked' or not jobs_after[0]['receipt'] or len(sinks) != 1:
                    raise RuntimeError('actual worker/sink receipt observation missing')
                acceptance = json.loads(next(iter(sinks.values()))['content'])
                if acceptance['business_application_id'] != jobs_after[0]['receipt'] or acceptance['payload']['item'] != jobs_after[0]['item']:
                    raise RuntimeError('observed sink/job business relation mismatch')
                if after_worker['outcome'] != 'UNKNOWN':
                    raise RuntimeError('snapshots manufactured causal completion')
                for field in ('execution_id', 'source_manifest_ref', 'runtime_ref', 'build_ref', 'window_id'):
                    changed = copy.deepcopy(worker); changed[field] = 'foreign'
                    try:
                        assess(changed, scope, worker_ref)
                    except ValueError:
                        continue
                    raise RuntimeError('altered stage provenance accepted')
                report = {**scope, 'pending_ref': pending_ref, 'worker_ref': worker_ref,
                    'before_worker': before_worker, 'after_worker': after_worker,
                    'launch_ref': fixture.store(run_dir, 'launch.json', app.launch), 'runtime': runtime,
                    'implementation_ref': impl_ref, 'role_isolation_ref': role_ref, 'outcome': 'UNKNOWN'}
                fixture.store(run_dir, 'capture-report.json', report)
                reports.append(report)
        if builds[0] != builds[1] or reports[0]['after_worker']['observed_queue_after'] != reports[1]['after_worker']['observed_queue_after']:
            raise RuntimeError('pinned input/observed queue reproduction changed')
        results_ref = fixture.store(out, 'control-result.json', {'side': 'SOURCE', 'runs': reports, 'reproduction_passed': True})
        q = {'schema_version': '1.0', 'qualification_id': 'ci3_async_completion_attempt_1', 'project_id': 'ci3_reference',
             'subject_kind': 'ENGINE', 'implementation_ref': impl_ref,
             'schema_refs': [fixture.ref((fixture.REPO / path).read_bytes()) for path in
                             ('specs/observation/causality.schema.json', 'specs/observation/effect-occurrence.schema.json', 'specs/schemas/trace-schema.json')],
             'capability_ids': ['source-async-causal-completion'], 'input_profile_ref': profile_ref,
             'control_suite_ref': suite_ref, 'control_result_refs': [results_ref], 'status': 'INCONCLUSIVE',
             'reason_codes': list(GAPS), 'limitations': ['Existing raw stages only; no qualified due/horizon or causal completion',
                'DB intermediate visibility and async causal/completion ABI remain PARTIAL', 'No witness/occurrence/trace publication']}
        print(fixture.store(out, 'qualification.json', q), flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    run(parser.parse_args().output)
