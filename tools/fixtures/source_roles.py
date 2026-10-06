#!/usr/bin/env python3
"""Qualify mediated SOURCE capture handles; no observation-window publication."""
import argparse
import copy
import json
import shutil
from pathlib import Path
import tempfile
import uuid

import ci3_reference as fixture


class SourceCapture:
    """A source-bound read capability. No caller endpoint, model or write port."""
    __slots__ = ('__read', '__source_ref')

    def __init__(self, read, source_ref):
        self.__read = read
        self.__source_ref = source_ref

    def capture(self, request):
        if set(request) != {'side', 'source_manifest_ref'} or request['side'] != 'SOURCE' or request['source_manifest_ref'] != self.__source_ref:
            raise ValueError('SOURCE capture scope contamination')
        observed = self.__read()
        if set(observed) != {'side', 'database', 'source_files', 'sink_files'} or observed['side'] != 'SOURCE':
            raise ValueError('SOURCE capture provenance contamination')
        # Serializing detaches values from the observer's allocation. Callers only
        # receive immutable bytes, never the runtime or a store/sink handle.
        return fixture.canonical(observed)


class ActualReader:
    """Checker handle: immutable actual bytes only, no expectation fallback."""
    __slots__ = ('__actual',)

    def __init__(self, actual):
        self.__actual = bytes(actual)

    def read(self):
        return json.loads(self.__actual)


FORBIDDEN = ('target_model_ref', 'target_artifact_ref', 'target_execution_ref',
             'comparator_expectations', 'pairwise_mapping', 'plan_ref', 'approval_ref',
             'expected_effects', 'expected_principal')


def controls(capture, source_ref):
    """Execute controls against the same capture implementation used by callers."""
    request = {'side': 'SOURCE', 'source_manifest_ref': source_ref}
    actual = capture.capture(request)
    results = []

    def check(name, passed):
        results.append({'id': name, 'passed': bool(passed)})
        if not passed:
            raise RuntimeError('role qualification failed: ' + name)

    check('source-only-capture', json.loads(actual)['side'] == 'SOURCE')
    for key in FORBIDDEN:
        rejected = False
        try:
            capture.capture({**request, key: {'poisoned': 'different expected owner and INSERT'}})
        except (ValueError, TypeError):
            rejected = True
        check('reject-' + key, rejected)
        check('unchanged-after-' + key, capture.capture(request) == actual)
    for request_mutant in ({'side': 'TARGET', 'source_manifest_ref': source_ref},
                           {'side': 'SOURCE', 'source_manifest_ref': 'sha256:' + '0' * 64}):
        rejected = False
        try:
            capture.capture(request_mutant)
        except ValueError:
            rejected = True
        check('reject-' + ('wrong-side' if request_mutant['side'] == 'TARGET' else 'stale-source'), rejected)
    # A deliberately poisoned model allocation is a negative-control input,
    # never executed as Behavior IR and never supplied to the source observer.
    shadow = copy.deepcopy(json.loads(actual))
    shadow['database'] += '\n999\tpoisoned-expected-insert\n'
    shadow['source_files'] = {'/state/poisoned': {'content': 'expected', 'mode': '777'}}
    reader = ActualReader(actual)
    check('shadow-allocation-does-not-alias-source', capture.capture(request) == actual)
    check('missing-write-remains-absent', 'poisoned-expected-insert' not in reader.read()['database'])
    detached = reader.read()
    detached['database'] = shadow['database']
    check('checker-read-does-not-alias-capture', fixture.canonical(reader.read()) == actual)
    for role, handle in (('observer', capture), ('checker', reader)):
        for method in ('write', 'execute', 'emit', 'set_state', 'transaction', 'restore', 'request'):
            check(role + '-has-no-' + method + '-handle', not hasattr(handle, method))
    check('no-measured-state-mutation', capture.capture(request) == actual)
    return actual, results


def required_controls():
    names = ['source-only-capture']
    for key in FORBIDDEN:
        names.extend(['reject-' + key, 'unchanged-after-' + key])
    names.extend(['reject-wrong-side', 'reject-stale-source',
                  'shadow-allocation-does-not-alias-source', 'missing-write-remains-absent',
                  'checker-read-does-not-alias-capture'])
    for role in ('observer', 'checker'):
        for method in ('write', 'execute', 'emit', 'set_state', 'transaction', 'restore', 'request'):
            names.append(role + '-has-no-' + method + '-handle')
    return names + ['no-measured-state-mutation', 'observer-db-principal-insert-denied', 'observer-server-write-audit']


def validate_results(results):
    names = [r['id'] for r in results]
    if len(names) != len(set(names)) or set(names) != set(required_controls()) or any(r['passed'] is not True for r in results):
        raise ValueError('Incomplete or failed role qualification controls')


def qualification(implementation_ref, schema_refs, profile_ref, suite_ref, results_ref):
    return {'schema_version': '1.0', 'qualification_id': 'ci3_source_roles_1',
            'project_id': 'ci3_reference', 'subject_kind': 'ENGINE',
            'implementation_ref': implementation_ref, 'schema_refs': schema_refs,
            'capability_ids': ['rwt-model-roles-1'], 'input_profile_ref': profile_ref,
            'control_suite_ref': suite_ref, 'control_result_refs': [results_ref],
            'limitations': ['Mediated SOURCE capture/checker handles only; same-account introspection is outside this boundary',
                            'No target execution, model interpreter, session codec or customer approval qualified',
                            'Old captures remain unqualified for this successor implementation',
                            'Controlled observation windows and witnessed trace publication not established',
                            'DB intermediate visibility and async causal/completion ABI remain PARTIAL'],
            'status': 'SUPPORTED', 'reason_codes': ['EXECUTED_SOURCE_ROLE_HANDLE_CONTROLS']}


def run(out):
    out.mkdir(mode=0o700)  # New output only; never modify retained evidence.
    with tempfile.TemporaryDirectory(prefix='rwt-source-role-snapshot-') as temporary:
        root = Path(temporary) / 'source'
        shutil.copytree(fixture.FIXTURE, root, ignore=shutil.ignore_patterns('evidence', 'README.md'))
        qualify(out, root)


def qualify(out, root):
    lock = fixture.pins(root)
    implementation = {'source_roles.py': fixture.ref(Path(__file__).read_bytes()),
                      'ci3_reference.py': fixture.ref(Path(fixture.__file__).read_bytes())}
    implementation_ref = fixture.store(out, 'implementation.json', implementation)
    manifest_ref = fixture.store(out, 'source-manifest.json', fixture.source_manifest(root))
    schemas = [fixture.REPO / 'specs/adapters/qualification.schema.json',
               fixture.REPO / 'specs/behavior/semantics.schema.json',
               fixture.REPO / 'specs/observation/execution-witness.schema.json']
    schema_refs = [fixture.ref(s.read_bytes()) for s in schemas]
    profile = {'side': 'SOURCE', 'fixture': 'ci3-reference-b1', 'source_manifest_ref': manifest_ref,
               'runtime_ref': fixture.ref((root / 'runtime.lock.json').read_bytes()),
               'implementation_ref': implementation_ref, 'python_version': lock['python_version'],
               'boundary': 'source capture/checker mediated handles; no measured-write or expected-value port'}
    profile_ref = fixture.store(out, 'input-profile.json', profile)
    suite_ref = fixture.store(out, 'control-suite.json', {'owner': 'BEHAVIOR-MODEL execution roles; X61',
                                                         'implementation_ref': implementation_ref,
                                                         'forbidden_capture_inputs': list(FORBIDDEN),
                                                         'required_controls': [{'id': name, 'passed': True} for name in required_controls()]})
    image, build_ref, runtime = fixture.build(lock, 'rewrotron-ci3:roles-build', out, root)
    execution_id = 'source_roles_' + uuid.uuid4().hex
    with fixture.SourceRuntime(image, lock, execution_id, root) as app:
        # Only the trusted harness owns preparation/runtime/driver capabilities.
        # It delivers a real health request before checking read-only handles.
        response = fixture.Driver(app).request('/health')
        if response['status'] != 200:
            raise RuntimeError('source application failed to execute')
        observer = fixture.Observer(app)
        capture = SourceCapture(observer.snapshot, manifest_ref)
        before, checks = controls(capture, manifest_ref)
        # Independently exercise the actual DB principal, not just Python API shape.
        denied = fixture.docker('exec', app.db, 'mysql', '-uobserver', '-psynthetic-observer-only',
                                '--batch', '-e', "INSERT INTO fixture.items(owner,name) VALUES(1,'poisoned-expected-insert')", check=False)
        if denied.returncode == 0 or b"INSERT command denied to user 'observer'" not in denied.stderr or capture.capture({'side': 'SOURCE', 'source_manifest_ref': manifest_ref}) != before:
            raise RuntimeError('observer write containment failed')
        checks.append({'id': 'observer-db-principal-insert-denied', 'passed': True})
        # The trusted supervisor reads the server audit independently of the
        # observer. A hidden write-then-restore cannot be excused by equal snapshots.
        audit = fixture.docker('exec', app.db, 'mysql', '-uroot', '-psynthetic-root-only',
                               '--batch', '--skip-column-names', '-e',
                               "SELECT argument FROM mysql.general_log WHERE user_host LIKE 'observer%' AND command_type='Query' ORDER BY event_time,thread_id").stdout
        queries = audit.decode().splitlines()
        (out / 'observer-server-audit.tsv').write_bytes(audit)
        writes = [query for query in queries if not query.lstrip().upper().startswith('SELECT ')]
        expected_write = "INSERT INTO fixture.items(owner,name) VALUES(1,'poisoned-expected-insert')"
        if not queries or writes != [expected_write]:
            raise RuntimeError('observer server write audit failed; actual queries retained')
        (out / 'observer-server-audit.tsv').write_bytes(audit)
        checks.append({'id': 'observer-server-write-audit', 'passed': True})
        validate_results(checks)
        launch_ref = fixture.store(out, 'launch.json', app.launch)
        actual_ref = fixture.store(out, 'capture.json', json.loads(before))
        result = {'side': 'SOURCE', 'execution_id': execution_id, 'fixture': 'ci3-reference-b1',
                  'source_manifest_ref': manifest_ref, 'runtime_ref': profile['runtime_ref'],
                  'build_ref': build_ref, 'loaded_image_ref': app.launch['Image'], 'launch_ref': launch_ref,
                  'implementation_ref': implementation_ref, 'profile_ref': profile_ref, 'control_suite_ref': suite_ref,
                  'capture_ref': actual_ref, 'runtime': runtime, 'controls': checks,
                  'observer_server_audit_ref': fixture.ref(audit),
                  'db_write_refusal': {'exit_code': denied.returncode,
                                       'stderr_ref': fixture.ref(denied.stderr)},
                  'source_health_response': response}
        (out / 'db-write-refusal.stderr').write_bytes(denied.stderr)
        results_ref = fixture.store(out, 'control-result.json', result)
    q = qualification(implementation_ref, schema_refs, profile_ref, suite_ref, results_ref)
    q_ref = fixture.store(out, 'qualification.json', q)
    fixture.store(out, 'artifact-index.json', {'side': 'SOURCE', 'qualification_ref': q_ref,
                                             'control_result_ref': results_ref, 'implementation_ref': implementation_ref})
    print(q_ref, flush=True)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('output', type=Path)
    run(parser.parse_args().output)
