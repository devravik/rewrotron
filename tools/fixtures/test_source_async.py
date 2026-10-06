"""Raw asynchronous stages never substitute for the missing X12 ABI."""
import copy
import json
from pathlib import Path
import unittest
from unittest.mock import patch

import ci3_reference as fixture
import source_async as asynchronous
import source_windows as windows


class AsyncStageTests(unittest.TestCase):
    def setUp(self):
        self.scope = {'side': 'SOURCE', 'execution_id': 'source_test', 'source_manifest_ref': 'sha256:'+'a'*64,
                      'runtime_ref': 'sha256:'+'b'*64, 'build_ref': 'sha256:'+'c'*64}
        observed = {'event_scheduler': 'ON', 'events': 0, 'routines': 0, 'event_privileges': 0,
                    'global_event_privileges': 0, 'jobs': [{'id': 1, 'item': 1, 'status': 'pending', 'receipt': None}]}
        raw = fixture.canonical(observed).decode()
        self.state = {'actual_state': {'side': 'SOURCE', 'sink_files': []}, 'queue': {
            'side': 'SOURCE', 'capture_method': 'fixed-read-only-server-census', 'observed': observed,
            'raw': raw, 'raw_ref': fixture.ref(raw.encode())}}
        observer = windows.Windows(self.scope, lambda: copy.deepcopy(self.state), 'sha256:'+'d'*64)
        with patch.object(windows.time, 'monotonic_ns', side_effect=[0, 10, 1000100, 2000100, 3000100]), \
             patch.object(windows.time, 'time_ns', return_value=1700000000000000000):
            _, identity = observer.invoke('HTTP /create', lambda: {'status': 202})
        self.record = json.loads(observer.records[identity]); self.ref = fixture.ref(fixture.canonical(self.record))

    def test_http_success_and_queue_registration_are_unknown(self):
        outcome = asynchronous.assess(self.record, self.scope, self.ref)
        self.assertEqual(outcome['outcome'], 'UNKNOWN')
        self.assertIsNone(outcome['schedule_horizon_ms'])
        self.assertIsNone(outcome['due_beyond_horizon'])
        self.assertEqual(outcome['async_causal_completion_abi'], 'PARTIAL')
        self.assertEqual(outcome['db_intermediate_visibility'], 'PARTIAL')

    def test_stored_receipt_and_worker_exit_do_not_create_causal_ack(self):
        changed = copy.deepcopy(self.record)
        changed['result'] = 0
        evidence = changed['post_capture']['queue']
        evidence['observed']['jobs'][0].update(status='acked', receipt='actual-receipt')
        evidence['raw'] = fixture.canonical(evidence['observed']).decode()
        evidence['raw_ref'] = fixture.ref(evidence['raw'].encode())
        outcome = asynchronous.assess(changed, self.scope, fixture.ref(fixture.canonical(changed)))
        self.assertEqual(outcome['outcome'], 'UNKNOWN')
        self.assertIn('missing-worker-parent-span', outcome['reason_codes'])

    def test_defaults_fabricated_completion_foreign_scope_rejected(self):
        for key in ('target_model_ref', 'target_artifact_ref', 'plan_ref', 'approval_ref', 'comparator', 'causal_ack'):
            changed = copy.deepcopy(self.record); changed[key] = 'invented'
            with self.assertRaises(ValueError):
                asynchronous.assess(changed, self.scope, fixture.ref(fixture.canonical(changed)))
        changed = copy.deepcopy(self.record)
        changed['observation_window'].update(schedule_horizon_ms=0, due_beyond_horizon=False)
        with self.assertRaises(ValueError):
            asynchronous.assess(changed, self.scope, fixture.ref(fixture.canonical(changed)))

    def test_wrong_execution_source_runtime_build_window_hash_rejected(self):
        for field in ('execution_id', 'source_manifest_ref', 'runtime_ref', 'build_ref', 'window_id'):
            changed = copy.deepcopy(self.record); changed[field] = 'foreign'
            with self.assertRaises(ValueError):
                asynchronous.assess(changed, self.scope, self.ref)

    def test_missing_census_or_fabricated_queue_is_not_evidence(self):
        changed = copy.deepcopy(self.record); changed['post_capture'].pop('queue')
        with self.assertRaises(ValueError):
            asynchronous.assess(changed, self.scope, fixture.ref(fixture.canonical(changed)))
        changed = copy.deepcopy(self.record); changed['post_capture']['queue']['observed']['jobs'] = []
        with self.assertRaises(ValueError):
            asynchronous.assess(changed, self.scope, fixture.ref(fixture.canonical(changed)))

    def test_repeat_assessment_is_deterministic_and_readonly(self):
        original = fixture.canonical(self.record)
        a = asynchronous.assess(self.record, self.scope, self.ref)
        self.assertEqual(a, asynchronous.assess(self.record, self.scope, self.ref))
        self.assertEqual(original, fixture.canonical(self.record))


class FreshAsyncEvidenceTests(unittest.TestCase):
    def test_fresh_stages_remain_unknown_and_cannot_cross_link_runs(self):
        root = fixture.FIXTURE / 'evidence/runs/async-horizon-b1'
        objects = {fixture.ref(fixture.canonical(json.loads(p.read_bytes()))): json.loads(p.read_bytes())
                   for p in root.rglob('*.json')}
        q = json.loads((root/'qualification.json').read_bytes())
        self.assertEqual(q['status'],'INCONCLUSIVE')
        self.assertEqual(objects[q['implementation_ref']]['source_async.py'],fixture.ref(Path(asynchronous.__file__).read_bytes()))
        runs = objects[q['control_result_refs'][0]]['runs']
        for run in runs:
            scope = {k:run[k] for k in ('side','execution_id','source_manifest_ref','runtime_ref','build_ref')}
            for field in ('pending_ref','worker_ref'):
                result = asynchronous.assess(objects[run[field]],scope,run[field])
                self.assertEqual(result['outcome'],'UNKNOWN')
                self.assertIsNone(result['schedule_horizon_ms'])
                self.assertIsNone(result['due_beyond_horizon'])
        self.assertEqual(runs[0]['build_ref'],runs[1]['build_ref'])
        self.assertNotEqual(runs[0]['execution_id'],runs[1]['execution_id'])
        scope = {k:runs[1][k] for k in ('side','execution_id','source_manifest_ref','runtime_ref','build_ref')}
        with self.assertRaises(ValueError):
            asynchronous.assess(objects[runs[0]['worker_ref']],scope,runs[0]['worker_ref'])


if __name__ == '__main__':
    unittest.main()
