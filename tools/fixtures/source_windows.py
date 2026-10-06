#!/usr/bin/env python3
"""Controlled SOURCE dispatch intervals and fresh raw CI3 capture; no traces."""
import argparse
import copy
from datetime import datetime, timezone
import json
from pathlib import Path
import shutil
import tempfile
import threading
import time
import types
import uuid

import ci3_reference as fixture
import source_roles as roles


def utc(ns):
    seconds, remainder = divmod(ns, 1000000000)
    return datetime.fromtimestamp(seconds, timezone.utc).strftime('%Y-%m-%dT%H:%M:%S') + '.%06dZ' % (remainder // 1000)


class Windows:
    """Supervisor-owned intervals. Callers cannot supply times or foreign scope."""
    def __init__(self, scope, capture, control_ref):
        if set(scope) != {'side', 'execution_id', 'source_manifest_ref', 'runtime_ref', 'build_ref'} or scope['side'] != 'SOURCE' or not scope['execution_id']:
            raise ValueError('invalid SOURCE timing scope')
        self.scope = copy.deepcopy(scope)
        self.capture = capture
        self.control_ref = control_ref
        self.records = {}
        self.visits = {}
        self.allocated = set()
        self.lock = threading.Lock()

    def invoke(self, operation_id, operation):
        if not isinstance(operation_id, str) or not operation_id:
            raise ValueError('missing operation identity')
        with self.lock:
            visit = self.visits.get(operation_id, 0) + 1
            self.visits[operation_id] = visit
            identity = 'source_window_' + uuid.uuid4().hex
            if identity in self.allocated:
                raise ValueError('duplicate window identity')
            self.allocated.add(identity)
        record = {**self.scope, 'window_id': identity, 'operation_id': operation_id, 'visit': visit,
                  'origin': 'ACTUAL_OBSERVED', 'status': 'INCOMPLETE', 'pre_capture': None,
                  'result': None, 'post_capture': None, 'anchor': None, 'events': [],
                  'observation_window': {'start': None, 'end': None, 'schedule_horizon_ms': None,
                                         'queue_state_ref': None, 'due_beyond_horizon': None,
                                         'control_refs': [self.control_ref]}}
        failure = None
        stage = 'pre_capture'
        try:
            record['pre_capture'] = self.capture()
            stage = 'clock'
            mono = time.monotonic_ns()
            wall = time.time_ns()
            bracket = time.monotonic_ns()
            if bracket < mono:
                raise ValueError('monotonic anchor reversed')
            record['anchor'] = {'monotonic_ns': mono, 'utc_ns': wall, 'sample_end_monotonic_ns': bracket}
            record['observation_window']['start'] = utc(wall)
            stage = 'execution'
            record['result'] = operation()
            observed = time.monotonic_ns()
            record['events'].append({'kind': 'operation_return', 'monotonic_ns': observed,
                                     'elapsed_ms': (observed - mono + 999999) // 1000000})
            stage = 'post_capture'
            record['post_capture'] = self.capture()
            observed = time.monotonic_ns()
            record['events'].append({'kind': 'post_capture', 'monotonic_ns': observed,
                                     'elapsed_ms': (observed - mono + 999999) // 1000000})
            record['status'] = 'INTERVAL_CAPTURED_COMPLETION_UNQUALIFIED'
        except BaseException as error:
            failure = error
            record['failure'] = {'stage': stage, 'error_type': type(error).__name__}
        finally:
            if record['anchor'] is not None:
                end = time.monotonic_ns()
                anchor = record['anchor']
                record['end_monotonic_ns'] = end
                record['timer_elapsed_ms'] = (end - anchor['monotonic_ns']) // 1000000
                record['observation_window']['end'] = utc(anchor['utc_ns'] + end - anchor['monotonic_ns'])
            with self.lock:
                # Detached immutable bytes keep concurrent consumers from editing records.
                self.records[identity] = fixture.canonical(record)
        if failure is not None:
            raise failure
        validate(record, self.scope)
        return record['result'], identity


def validate(record, scope, expected_ref=None):
    if expected_ref is not None and fixture.ref(fixture.canonical(record)) != expected_ref:
        raise ValueError('window artifact changed')
    if any(record.get(k) != v for k, v in scope.items()) or scope.get('side') != 'SOURCE':
        raise ValueError('window execution/source/runtime/build scope mismatch')
    required = {'window_id', 'operation_id', 'visit', 'origin', 'status', 'pre_capture', 'post_capture',
                'result', 'anchor', 'events', 'observation_window', 'end_monotonic_ns', 'timer_elapsed_ms'} | set(scope)
    if set(record) != required or record['origin'] != 'ACTUAL_OBSERVED' or record['status'] != 'INTERVAL_CAPTURED_COMPLETION_UNQUALIFIED':
        raise ValueError('missing/incomplete/foreign window evidence')
    if record['pre_capture'] is None or record['post_capture'] is None or not record['window_id'] or record['visit'] < 1:
        raise ValueError('missing capture/identity')
    a = record['anchor']
    end = record['end_monotonic_ns']
    w = record['observation_window']
    if not a or set(a) != {'monotonic_ns', 'utc_ns', 'sample_end_monotonic_ns'} or set(w) != {'start', 'end', 'schedule_horizon_ms', 'queue_state_ref', 'due_beyond_horizon', 'control_refs'} or not isinstance(end, int) or end <= a['sample_end_monotonic_ns'] or a['sample_end_monotonic_ns'] < a['monotonic_ns']:
        raise ValueError('invalid controlled interval')
    if w['start'] != utc(a['utc_ns']) or w['end'] != utc(a['utc_ns'] + end - a['monotonic_ns']) or w['end'] <= w['start']:
        raise ValueError('forged/missing boundary')
    if record['timer_elapsed_ms'] != (end - a['monotonic_ns']) // 1000000 or not w['control_refs']:
        raise ValueError('unqualified clock projection')
    if w['schedule_horizon_ms'] is not None or w['due_beyond_horizon'] is not None or w['queue_state_ref'] is not None:
        raise ValueError('unqualified completion/queue claim')
    if [e['kind'] for e in record['events']] != ['operation_return', 'post_capture']:
        raise ValueError('missing capture event')
    if record['events'][0]['monotonic_ns'] > record['events'][1]['monotonic_ns']:
        raise ValueError('capture events reordered')
    for event in record['events']:
        if set(event) != {'kind', 'monotonic_ns', 'elapsed_ms'}:
            raise ValueError('foreign timing source')
        elapsed = event['monotonic_ns'] - a['monotonic_ns']
        if elapsed < 0 or event['monotonic_ns'] > end or event['elapsed_ms'] != (elapsed + 999999) // 1000000:
            raise ValueError('event outside interval/incorrect activity rounding')


def validate_collection(records, scope):
    seen, logical = set(), set()
    for record in records:
        validate(record, scope)
        key = (record['operation_id'], record['visit'])
        if record['window_id'] in seen or key in logical:
            raise ValueError('reused physical/logical interval')
        seen.add(record['window_id'])
        logical.add(key)


def validate_corpus(corpus, records, scope):
    validate_collection(records, scope)
    indexed = {fixture.ref(fixture.canonical(r)): r for r in records}
    if len(indexed) != len(records) or set(corpus['window_refs']) != set(indexed):
        raise ValueError('missing/reused raw window')
    seen = set()
    for observation in corpus['observations']:
        ref = observation['window_ref']
        if ref not in indexed or ref in seen:
            raise ValueError('missing/duplicated observation window binding')
        seen.add(ref)
        record = indexed[ref]
        expected_operation = 'HTTP ' + observation['request']['path'] if 'response' in observation else 'PROCESS worker'
        if observation['window_id'] != record['window_id'] or record['operation_id'] != expected_operation:
            raise ValueError('window attached to wrong operation')
        if 'response' in observation and record['result'] != observation['response']:
            raise ValueError('window response differs from actual capture')
        if record['pre_capture'] != observation['before'] or record['post_capture'] != observation['after']:
            raise ValueError('window state differs from actual capture')


def observe(app, out, execution_id, scope, control_ref):
    source = roles.SourceCapture(fixture.Observer(app).snapshot, scope['source_manifest_ref'])

    def snapshot():
        return json.loads(source.capture({'side': 'SOURCE', 'source_manifest_ref': scope['source_manifest_ref']}))

    windows = Windows(scope, snapshot, control_ref)
    driver = fixture.Driver(app)
    response_ids = {}
    worker_ids = []

    class TimedDriver:
        def __init__(self, runtime):
            if runtime is not app:
                raise ValueError('wrong runtime')

        def request(self, path, fields=None, cookie=None):
            response, window_id = windows.invoke('HTTP ' + path, lambda: driver.request(path, fields, cookie))
            response_ids[id(response)] = window_id  # Object association, never response equality/order.
            return response

        def worker(self):
            result, window_id = windows.invoke('PROCESS worker', driver.worker)
            worker_ids.append(window_id)
            return result

    class ReadObserver:
        def __init__(self, runtime):
            if runtime is not app:
                raise ValueError('wrong runtime')

        def snapshot(self):
            return snapshot()

    # Isolated callable dependency binding reuses the unchanged reviewed scenario.
    # The module globals/old source implementation are never modified or patched.
    scenario = types.FunctionType(fixture.observe.__code__,
                                  {**fixture.observe.__globals__, 'Driver': TimedDriver, 'Observer': ReadObserver})
    try:
        corpus = scenario(app, out, execution_id)
    finally:
        for key, raw in windows.records.items():
            (out / (key + '.json')).write_bytes(raw + b'\n')
    records = [json.loads(raw) for raw in windows.records.values()]
    validate_collection(records, scope)
    for observation in corpus['observations']:
        identity = response_ids[id(observation['response'])] if 'response' in observation else worker_ids[0]
        observation['window_ref'] = fixture.ref(windows.records[identity])
        observation['window_id'] = identity
    corpus['window_refs'] = [fixture.ref(windows.records[r['window_id']]) for r in records]
    for record in records:
        expected = fixture.ref(fixture.canonical(record))
        for field in ('execution_id', 'source_manifest_ref', 'runtime_ref', 'build_ref', 'window_id'):
            poisoned = copy.deepcopy(record)
            poisoned[field] = 'foreign'
            try:
                validate(poisoned, scope, expected)
            except ValueError:
                continue
            raise RuntimeError('altered window control accepted: ' + field)
        for field in ('start', 'end'):
            poisoned = copy.deepcopy(record)
            poisoned['observation_window'][field] = None
            try:
                validate(poisoned, scope)
            except ValueError:
                continue
            raise RuntimeError('missing boundary control accepted')
    corpus['qualification_limits'] = {'db_intermediate_visibility': 'PARTIAL', 'async_causal_completion_abi': 'PARTIAL'}
    validate_corpus(corpus, records, scope)
    return corpus, records


def run(out):
    out.mkdir(mode=0o700)
    role_root = fixture.FIXTURE / 'evidence/roles/run-0'
    role_raw = json.loads((role_root / 'qualification.json').read_bytes())
    role_ref = fixture.ref(fixture.canonical(role_raw))
    role_index = json.loads((role_root / 'artifact-index.json').read_bytes())
    if role_index['qualification_ref'] != role_ref:
        raise ValueError('role qualification digest changed')
    for filename, digest in [('implementation.json', role_raw['implementation_ref']),
                             ('input-profile.json', role_raw['input_profile_ref']),
                             ('control-suite.json', role_raw['control_suite_ref']),
                             ('control-result.json', role_raw['control_result_refs'][0])]:
        if fixture.ref(fixture.canonical(json.loads((role_root / filename).read_bytes()))) != digest:
            raise ValueError('role qualification closure changed')
    role_result = json.loads((role_root / 'control-result.json').read_bytes())
    roles.validate_results(role_result['controls'])
    # Refuse stale role bytes/profile rather than inferring role support.
    role_impl = json.loads((role_root / 'implementation.json').read_bytes())
    if role_raw['status'] != 'SUPPORTED' or role_raw['capability_ids'] != ['rwt-model-roles-1'] or role_impl != {
            'source_roles.py': fixture.ref(Path(roles.__file__).read_bytes()),
            'ci3_reference.py': fixture.ref(Path(fixture.__file__).read_bytes())}:
        raise ValueError('stale/unqualified source role implementation')
    implementation_ref = fixture.store(out, 'implementation.json', {
        'source_windows.py': fixture.ref(Path(__file__).read_bytes()), 'role_isolation_ref': role_ref,
        'ci3_reference.py': role_impl['ci3_reference.py'], 'source_roles.py': role_impl['source_roles.py']})
    controls_ref = fixture.store(out, 'clock-controls.json', {
        'implementation_ref': implementation_ref, 'clock': 'supervisor time.monotonic_ns anchored once to time.time_ns at dispatch',
        'activity_rounding': 'ceil milliseconds', 'timer_rounding': 'floor milliseconds',
        'completion': 'UNQUALIFIED; interval capture only'})
    with tempfile.TemporaryDirectory(prefix='rwt-window-source-') as temporary:
        root = Path(temporary) / 'source'
        shutil.copytree(fixture.FIXTURE, root, ignore=shutil.ignore_patterns('evidence', 'README.md'))
        lock = fixture.pins(root)
        source_ref = fixture.store(out, 'source-manifest.json', fixture.source_manifest(root))
        profile_ref = fixture.store(out, 'input-profile.json', {
            'side': 'SOURCE', 'fixture': 'ci3-reference-b1', 'source_manifest_ref': source_ref,
            'runtime_ref': fixture.ref((root / 'runtime.lock.json').read_bytes()),
            'implementation_ref': implementation_ref, 'role_isolation_ref': role_ref,
            'clock_controls_ref': controls_ref, 'boundary': 'raw dispatch-through-post-capture intervals only'})
        results, projections, build_refs = [], [], []
        for run_index in range(2):
            image, build_ref, runtime = fixture.build(lock, 'rewrotron-ci3:windows-' + str(run_index), out, root)
            build_refs.append(build_ref)
            execution_id = 'source_windows_' + uuid.uuid4().hex
            scope = {'side': 'SOURCE', 'execution_id': execution_id, 'source_manifest_ref': source_ref,
                     'runtime_ref': fixture.ref((root / 'runtime.lock.json').read_bytes()), 'build_ref': build_ref}
            run_dir = out / execution_id
            run_dir.mkdir()
            with fixture.SourceRuntime(image, lock, execution_id, root) as app:
                corpus, records = observe(app, run_dir, execution_id, scope, controls_ref)
                failure_probe = Windows(scope, lambda: json.loads(roles.SourceCapture(fixture.Observer(app).snapshot, source_ref).capture({'side': 'SOURCE', 'source_manifest_ref': source_ref})), controls_ref)
                try:
                    failure_probe.invoke('invalid-delivery-control', lambda: fixture.Driver(app).request('invalid-path'))
                except ValueError:
                    pass
                else:
                    raise RuntimeError('failed-delivery timing control did not fail')
                failure_record = json.loads(next(iter(failure_probe.records.values())))
                if failure_record['status'] != 'INCOMPLETE' or failure_record['observation_window']['end'] is None:
                    raise RuntimeError('failed-delivery interval fabricated success/missing end')
                failure_ref = fixture.store(run_dir, 'failed-delivery-control.json', failure_record)
                corpus_ref = fixture.store(run_dir, 'corpus.json', corpus)
                launch_ref = fixture.store(run_dir, 'launch.json', app.launch)
                result = {**scope, 'corpus_ref': corpus_ref, 'launch_ref': launch_ref,
                          'role_isolation_ref': role_ref, 'clock_controls_ref': controls_ref,
                          'implementation_ref': implementation_ref, 'profile_ref': profile_ref,
                          'window_refs': corpus['window_refs'], 'window_count': len(records), 'failed_delivery_control_ref': failure_ref,
                          'runtime': runtime, 'all_windows_valid': True}
                results.append(result)
                projections.append(corpus['reproduction_projection'])
                fixture.store(run_dir, 'capture-report.json', result)
            print('CAPTURED ' + execution_id, flush=True)
        if build_refs[0] != build_refs[1] or projections[0] != projections[1]:
            raise RuntimeError('pinned build/bounded response reproduction changed')
        suite_ref = fixture.store(out, 'control-suite.json', {'clock_controls_ref': controls_ref,
                                                             'required': ['live-boundaries', 'all-driver-operations-covered', 'unique-intervals', 'pinned-build-repetition', 'altered-window-rejected', 'missing-boundary-rejected', 'failed-delivery-incomplete']})
        results_ref = fixture.store(out, 'control-result.json', {'side': 'SOURCE', 'runs': results,
                                                                'control_suite_ref': suite_ref, 'reproduction_passed': True})
        schemas = [fixture.REPO / 'specs/schemas/trace-schema.json', fixture.REPO / 'specs/behavior/semantics.schema.json']
        q = {'schema_version': '1.0', 'qualification_id': 'ci3_source_windows_1', 'project_id': 'ci3_reference',
             'subject_kind': 'ENGINE', 'implementation_ref': implementation_ref,
             'schema_refs': [fixture.ref(s.read_bytes()) for s in schemas],
             'capability_ids': ['source-dispatch-interval-capture'], 'input_profile_ref': profile_ref,
             'control_suite_ref': suite_ref, 'control_result_refs': [results_ref], 'status': 'SUPPORTED',
             'reason_codes': ['EXECUTED_CONTROLLED_SOURCE_INTERVALS'],
             'limitations': ['Raw intervals only; no witness/occurrence/trace publication',
                             'No qualified horizon/queue/completion or business clock namespace',
                             'DB intermediate visibility and async causal/completion ABI remain PARTIAL']}
        fixture.store(out, 'qualification.json', q)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    run(parser.parse_args().output)
