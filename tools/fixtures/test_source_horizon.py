"""Defaults cannot substitute for positive Trace 2.1 horizon evidence."""
import copy
import json
from pathlib import Path
import unittest
from unittest.mock import patch

import ci3_reference as fixture
import source_horizon as horizon
import source_windows as windows


class HorizonTests(unittest.TestCase):
    def setUp(self):
        self.scope = {'side': 'SOURCE', 'execution_id': 'test_run', 'source_manifest_ref': 'sha256:' + 'a'*64,
                      'runtime_ref': 'sha256:' + 'b'*64, 'build_ref': 'sha256:' + 'c'*64}
        self.profile = {k: self.scope[k] for k in ('side', 'source_manifest_ref', 'runtime_ref', 'build_ref')}
        self.profile.update(operation_id='HTTP /health', completion=horizon.completion())
        observed = {'event_scheduler': 'OFF', 'events': 0, 'routines': 0, 'jobs': [], 'event_privileges': 0, 'global_event_privileges': 0}
        raw = fixture.canonical(observed).decode()
        self.capture = {'actual_state': {'side': 'SOURCE', 'database': 'actual readonly state'},
                        'queue': {'side': 'SOURCE', 'capture_method': 'fixed-read-only-server-census',
                                  'worker_ref': 'sha256:' + 'd'*64, 'raw': raw,
                                  'raw_ref': fixture.ref(raw.encode()), 'observed': observed}}
        observer = windows.Windows(self.scope, lambda: copy.deepcopy(self.capture), 'sha256:'+'e'*64)
        with patch.object(windows.time, 'monotonic_ns', side_effect=[0, 10, 1000100, 2000100, 3000100]), \
             patch.object(windows.time, 'time_ns', return_value=1700000000000000000):
            _, identity = observer.invoke('HTTP /health', lambda: {'status': 200})
        self.record = json.loads(observer.records[identity])
        self.ref = fixture.ref(fixture.canonical(self.record))

    def test_positive_synchronous_evidence_not_not_applicable(self):
        window, proof = horizon.derive(self.record, self.scope, self.ref, self.profile)
        self.assertEqual(window['schedule_horizon_ms'], 0)
        self.assertIs(window['due_beyond_horizon'], False)
        self.assertEqual(window['queue_state_ref'], fixture.ref(fixture.canonical(proof)))
        self.assertIn('async completion UNAVAILABLE', proof['completion_claim'])

    def test_default_values_cannot_replace_missing_evidence(self):
        for field in ('queue', 'actual_state'):
            bad = copy.deepcopy(self.record)
            bad['post_capture'].pop(field)
            with self.assertRaises(ValueError):
                horizon.derive(bad, self.scope, fixture.ref(fixture.canonical(bad)), self.profile)
        bad = copy.deepcopy(self.record)
        bad['observation_window'].update(schedule_horizon_ms=0, due_beyond_horizon=False)
        with self.assertRaises(ValueError):
            horizon.derive(bad, self.scope, fixture.ref(fixture.canonical(bad)), self.profile)

    def test_required_horizon_and_due_status_missing_or_not_applicable_rejected(self):
        for field in ('schedule_horizon_ms', 'due_beyond_horizon'):
            for value in ('missing', 'NOT_APPLICABLE'):
                bad = copy.deepcopy(self.record)
                if value == 'missing':
                    bad['observation_window'].pop(field)
                else:
                    bad['observation_window'][field] = value
                with self.assertRaises(ValueError):
                    horizon.derive(bad, self.scope, fixture.ref(fixture.canonical(bad)), self.profile)

    def test_pending_delayed_or_unknown_scheduler_cannot_qualify(self):
        for field, value in [('jobs', [{'id': 1, 'status': 'pending'}]), ('events', 1),
                             ('routines', 1), ('event_scheduler', 'UNKNOWN'), ('event_privileges', 1), ('jobs', None)]:
            bad = copy.deepcopy(self.record)
            evidence = bad['post_capture']['queue']
            evidence['observed'][field] = value
            evidence['raw'] = fixture.canonical(evidence['observed']).decode()
            evidence['raw_ref'] = fixture.ref(evidence['raw'].encode())
            with self.assertRaises(ValueError):
                horizon.derive(bad, self.scope, fixture.ref(fixture.canonical(bad)), self.profile)

    def test_altered_execution_window_runtime_and_source_rejected(self):
        for field in ('execution_id', 'window_id', 'runtime_ref', 'source_manifest_ref', 'build_ref'):
            bad = copy.deepcopy(self.record); bad[field] = 'wrong'
            with self.assertRaises(ValueError):
                horizon.derive(bad, self.scope, self.ref, self.profile)

    def test_foreign_profile_or_async_completion_not_qualified(self):
        for field in ('target_model_ref', 'plan_ref', 'approval_ref', 'comparator_expectations'):
            bad = {**self.profile, field: 'poison'}
            with self.assertRaises(ValueError):
                horizon.derive(self.record, self.scope, self.ref, bad)
        bad = copy.deepcopy(self.profile); bad['completion']['mode'] = 'SINK_ACK_QUIESCENCE'
        with self.assertRaises(ValueError):
            horizon.derive(self.record, self.scope, self.ref, bad)

    def test_queue_tampering_and_missing_due_semantics_fail(self):
        bad = copy.deepcopy(self.record)
        bad['post_capture']['queue']['observed']['jobs'] = [{'id': 1}]
        with self.assertRaises(ValueError):
            horizon.derive(bad, self.scope, fixture.ref(fixture.canonical(bad)), self.profile)
        bad = copy.deepcopy(self.record); bad['post_capture']['queue']['observed'].pop('event_scheduler')
        with self.assertRaises(ValueError):
            horizon.derive(bad, self.scope, fixture.ref(fixture.canonical(bad)), self.profile)

    def test_reproducible_projection_preserves_raw_interval(self):
        before = fixture.canonical(self.record)
        a = horizon.derive(self.record, self.scope, self.ref, self.profile)
        b = horizon.derive(self.record, self.scope, self.ref, self.profile)
        self.assertEqual(a, b)
        self.assertEqual(before, fixture.canonical(self.record))


class RuntimeHorizonTests(unittest.TestCase):
    def test_qualified_runtime_closes_profile_hashes_and_pending_control(self):
        root = fixture.FIXTURE / 'evidence/runs/health-horizon-b1'
        objects = {fixture.ref(fixture.canonical(json.loads(p.read_bytes()))): json.loads(p.read_bytes())
                   for p in root.rglob('*.json')}
        q = json.loads((root / 'qualification.json').read_bytes())
        impl = objects[q['implementation_ref']]
        self.assertEqual(impl['source_horizon.py'], fixture.ref(Path(horizon.__file__).read_bytes()))
        runs = objects[q['control_result_refs'][0]]['runs']
        for run in runs:
            scope = {k:run[k] for k in ('side','execution_id','source_manifest_ref','runtime_ref','build_ref')}
            window, proof = horizon.derive(objects[run['raw_window_ref']], scope, run['raw_window_ref'], run['profile'])
            self.assertEqual(window, run['observation_window'])
            self.assertEqual(objects[window['queue_state_ref']], proof)
            with self.assertRaises(ValueError):
                horizon.derive(objects[run['pending_control_ref']], scope, run['pending_control_ref'], run['profile'])
            self.assertEqual(run['db_intermediate_visibility'],'PARTIAL')
            self.assertEqual(run['async_causal_completion_abi'],'PARTIAL')
            self.assertEqual(run['implementation_ref'],q['implementation_ref'])
        self.assertEqual(runs[0]['build_ref'], runs[1]['build_ref'])
        self.assertEqual(runs[0]['profile'], runs[1]['profile'])
        self.assertEqual(objects[runs[0]['raw_window_ref']]['pre_capture'],objects[runs[1]['raw_window_ref']]['pre_capture'])
        self.assertNotEqual(runs[0]['execution_id'],runs[1]['execution_id'])


if __name__ == '__main__':
    unittest.main()
