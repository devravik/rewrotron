#!/usr/bin/env python3
"""Qualify the Trace 2.1 horizon of the pinned synchronous /health scope only."""
import argparse
import copy
import json
from pathlib import Path
import shutil
import tempfile
import uuid

import ci3_reference as fixture
import source_roles as roles
import source_windows as windows


def completion():
    return {'mode': 'HTTP_RESPONSE', 'timeout_ms': 5000, 'expected_receipts': 0,
            'quiescence_ms': 0, 'predicate': {'op': 'literal', 'value_type': 'boolean', 'value': True}}


def queue(app):
    # Trusted supervisor diagnostics, never a checker or application write port.
    query = "SELECT JSON_OBJECT('event_scheduler',@@event_scheduler,'events',(SELECT COUNT(*) FROM information_schema.EVENTS WHERE EVENT_SCHEMA='fixture'),'event_privileges',(SELECT COUNT(*) FROM mysql.db WHERE Db='fixture' AND User='fixture' AND Event_priv='Y'),'global_event_privileges',(SELECT COUNT(*) FROM mysql.user WHERE User='fixture' AND Event_priv='Y'),'routines',(SELECT COUNT(*) FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA='fixture'),'jobs',(SELECT COALESCE(JSON_ARRAYAGG(JSON_OBJECT('id',id,'item',item,'status',status,'receipt',receipt)),JSON_ARRAY()) FROM fixture.jobs));"
    raw = fixture.docker('exec', app.db, 'mysql', '-uroot', '-psynthetic-root-only',
                         '--batch', '--raw', '--skip-column-names', '-e', query).stdout
    return {'side': 'SOURCE', 'capture_method': 'fixed-read-only-server-census',
            'worker_ref': fixture.ref((app.root / 'worker.php').read_bytes()),
            'raw_ref': fixture.ref(raw), 'raw': raw.decode(), 'observed': json.loads(raw)}


def derive(record, scope, expected_ref, profile):
    # A structurally valid 0/false window is insufficient: require original actual
    # interval/hash/scope plus positive scheduler and queue capture in both phases.
    windows.validate(record, scope, expected_ref)
    if set(profile) != {'side', 'source_manifest_ref', 'runtime_ref', 'build_ref', 'operation_id', 'completion'} or profile != {
            'side': 'SOURCE', 'source_manifest_ref': scope['source_manifest_ref'],
            'runtime_ref': scope['runtime_ref'], 'build_ref': scope['build_ref'],
            'operation_id': 'HTTP /health', 'completion': completion()}:
        raise ValueError('unqualified completion/source/runtime/build profile')
    if record['operation_id'] != profile['operation_id'] or record['result']['status'] != 200 or record['timer_elapsed_ms'] >= completion()['timeout_ms']:
        raise ValueError('wrong operation or incomplete synchronous return')
    for phase in ('pre_capture', 'post_capture'):
        capture = record[phase]
        if set(capture) != {'actual_state', 'queue'} or capture['actual_state']['side'] != 'SOURCE':
            raise ValueError('missing source/queue capture')
        evidence = capture['queue']
        if set(evidence) != {'side', 'capture_method', 'worker_ref', 'raw_ref', 'raw', 'observed'} or evidence['side'] != 'SOURCE' or evidence['capture_method'] != 'fixed-read-only-server-census':
            raise ValueError('foreign/unqualified scheduler capture')
        if fixture.ref(evidence['raw'].encode()) != evidence['raw_ref'] or json.loads(evidence['raw']) != evidence['observed']:
            raise ValueError('altered queue capture')
        observed = evidence['observed']
        if set(observed) != {'event_scheduler', 'events', 'routines', 'jobs', 'event_privileges', 'global_event_privileges'} or observed['event_scheduler'] not in ('OFF', 'ON') or observed['events'] != 0 or observed['routines'] != 0 or observed['event_privileges'] != 0 or observed['global_event_privileges'] != 0 or observed['jobs'] != []:
            raise ValueError('pending/unsupported scheduled work; horizon unavailable')
    if record['pre_capture']['actual_state'] != record['post_capture']['actual_state']:
        raise ValueError('unexpected state mutation in qualified readonly health scope')
    proof = {**scope, 'raw_window_ref': expected_ref, 'window_id': record['window_id'],
             'profile': profile, 'origin': 'ACTUAL_OBSERVED',
             'pre_queue': record['pre_capture']['queue'], 'post_queue': record['post_capture']['queue'],
             'semantic_origin': 'BEHAVIOR-MODEL proven HTTP_RESPONSE synchronous closure; no scheduler/pending jobs in pinned /health scope',
             'completion_claim': 'HTTP_RESPONSE only; async completion UNAVAILABLE'}
    result = copy.deepcopy(record['observation_window'])
    result.update(schedule_horizon_ms=0, due_beyond_horizon=False,
                  queue_state_ref=fixture.ref(fixture.canonical(proof)))
    return result, proof


def run(out):
    out.mkdir(mode=0o700)
    role_root = fixture.FIXTURE / 'evidence/roles/run-0'
    role = json.loads((role_root / 'qualification.json').read_bytes())
    role_ref = fixture.ref(fixture.canonical(role))
    role_index = json.loads((role_root / 'artifact-index.json').read_bytes())
    impl = json.loads((role_root / 'implementation.json').read_bytes())
    if role_index['qualification_ref'] != role_ref or role['status'] != 'SUPPORTED' or impl != {
            'source_roles.py': fixture.ref(Path(roles.__file__).read_bytes()),
            'ci3_reference.py': fixture.ref(Path(fixture.__file__).read_bytes())}:
        raise ValueError('stale role qualification')
    implementation_ref = fixture.store(out, 'implementation.json', {
        'source_horizon.py': fixture.ref(Path(__file__).read_bytes()),
        'source_windows.py': fixture.ref(Path(windows.__file__).read_bytes()),
        'source_roles.py': impl['source_roles.py'], 'ci3_reference.py': impl['ci3_reference.py']})
    completion_ref = fixture.store(out, 'completion.json', completion())
    controls_ref = fixture.store(out, 'control-suite.json', {'implementation_ref': implementation_ref,
        'required': ['source-health-success', 'positive-empty-queue-census', 'positive-no-events-and-no-event-privileges', 'profile-horizon-origin',
                     'missing-census-rejected', 'altered-field-rejected', 'real-pending-job-rejected', 'independent-build-reproduction'],
        'boundary': 'HTTP /health only; no async completion/causal qualification'})
    results, builds = [], []
    with tempfile.TemporaryDirectory(prefix='rwt-horizon-sealed-') as temporary:
        root = Path(temporary) / 'source'
        shutil.copytree(fixture.FIXTURE, root, ignore=shutil.ignore_patterns('evidence', 'README.md'))
        lock = fixture.pins(root)
        source_ref = fixture.store(out, 'source-manifest.json', fixture.source_manifest(root))
        runtime_ref = fixture.ref((root / 'runtime.lock.json').read_bytes())
        input_profile_ref = fixture.store(out, 'input-profile.json', {'side': 'SOURCE', 'fixture': 'ci3-reference-b1',
            'source_manifest_ref': source_ref, 'runtime_ref': runtime_ref, 'implementation_ref': implementation_ref,
            'completion_ref': completion_ref, 'role_isolation_ref': role_ref, 'operation_id': 'HTTP /health'})
        for index in range(2):
            image, build_ref, actual_runtime = fixture.build(lock, 'rewrotron-ci3:horizon-' + str(index), out, root)
            builds.append(build_ref)
            execution_id = 'source_horizon_' + uuid.uuid4().hex
            run_dir = out / execution_id
            run_dir.mkdir()
            scope = {'side': 'SOURCE', 'execution_id': execution_id, 'source_manifest_ref': source_ref,
                     'runtime_ref': runtime_ref, 'build_ref': build_ref}
            profile = {k: scope[k] for k in ('side', 'source_manifest_ref', 'runtime_ref', 'build_ref')}
            profile.update(operation_id='HTTP /health', completion=completion())
            with fixture.SourceRuntime(image, lock, execution_id, root) as app:
                source = roles.SourceCapture(fixture.Observer(app).snapshot, source_ref)
                def snapshot():
                    return {'actual_state': json.loads(source.capture({'side': 'SOURCE', 'source_manifest_ref': source_ref})), 'queue': queue(app)}
                observer = windows.Windows(scope, snapshot, controls_ref)
                driver = fixture.Driver(app)
                _, identity = observer.invoke('HTTP /health', lambda: driver.request('/health'))
                record = json.loads(observer.records[identity])
                raw_ref = fixture.store(run_dir, 'raw-window.json', record)
                window, proof = derive(record, scope, raw_ref, profile)
                proof_ref = fixture.store(run_dir, 'queue-proof.json', proof)
                assert window['queue_state_ref'] == proof_ref
                fixture.store(run_dir, 'observation-window.json', window)
                for field in ('execution_id', 'runtime_ref', 'source_manifest_ref', 'build_ref'):
                    bad = copy.deepcopy(record); bad[field] = 'foreign'
                    try:
                        derive(bad, scope, raw_ref, profile)
                    except ValueError:
                        continue
                    raise RuntimeError('foreign horizon binding accepted')
                bad = copy.deepcopy(record); bad['post_capture'].pop('queue')
                try:
                    derive(bad, scope, fixture.ref(fixture.canonical(bad)), profile)
                except ValueError:
                    pass
                else:
                    raise RuntimeError('missing required census accepted')
                # The application alone creates the negative-control pending job.
                login = driver.request('/login', {'name': 'alice', 'password': 'fixture-password'})
                cookie = fixture.cookie(login)
                token = fixture.body(driver.request('/csrf', cookie=cookie))['token']
                created = driver.request('/create', {'name': 'horizon-negative', 'csrf': token}, cookie)
                if created['status'] != 202:
                    raise RuntimeError('pending-job control not executed')
                _, pending_identity = observer.invoke('HTTP /health', lambda: driver.request('/health'))
                pending = json.loads(observer.records[pending_identity])
                pending_ref = fixture.store(run_dir, 'pending-job-control.json', pending)
                try:
                    derive(pending, scope, pending_ref, profile)
                except ValueError:
                    pass
                else:
                    raise RuntimeError('pending async work falsely qualified as synchronous')
                report = {**scope, 'raw_window_ref': raw_ref, 'window_id': identity,
                    'observation_window': window, 'pending_control_ref': pending_ref,
                    'profile': profile, 'implementation_ref': implementation_ref, 'role_isolation_ref': role_ref,
                    'launch_ref': fixture.store(run_dir, 'launch.json', app.launch), 'runtime': actual_runtime,
                    'outcome': 'SUPPORTED_SYNCHRONOUS_HEALTH_ONLY',
                    'db_intermediate_visibility': 'PARTIAL', 'async_causal_completion_abi': 'PARTIAL'}
                results.append(report)
                fixture.store(run_dir, 'capture-report.json', report)
        if builds[0] != builds[1]:
            raise RuntimeError('independent build identity changed')
        results_ref = fixture.store(out, 'control-result.json', {'side': 'SOURCE', 'runs': results, 'reproduction_passed': True})
        q = {'schema_version': '1.0', 'qualification_id': 'ci3_health_horizon_1', 'project_id': 'ci3_reference',
             'subject_kind': 'ENGINE', 'implementation_ref': implementation_ref,
             'schema_refs': [fixture.ref((fixture.REPO / p).read_bytes()) for p in ('specs/schemas/trace-schema.json','specs/behavior/semantics.schema.json')],
             'capability_ids': ['source-synchronous-health-horizon'], 'input_profile_ref': input_profile_ref,
             'control_suite_ref': controls_ref, 'control_result_refs': [results_ref], 'status': 'SUPPORTED',
             'reason_codes': ['LIVE_QUEUE_CENSUS_AND_SYNCHRONOUS_CLOSURE'],
             'limitations': ['Only pinned /health with empty job queue, no native events/routines and no application event privileges',
                            'No retroactive qualification of windows-b1; scheduled/async scopes remain blocked',
                            'No witness/trace publication; DB intermediate and async causal/completion ABI remain PARTIAL']}
        print(fixture.store(out, 'qualification.json', q), flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    run(parser.parse_args().output)
