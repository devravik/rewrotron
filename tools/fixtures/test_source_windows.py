"""Timing attribution falsification; no model interpretation or target execution."""
import copy
from concurrent.futures import ThreadPoolExecutor
import json
from pathlib import Path
import unittest
from unittest.mock import patch

import ci3_reference as fixture
import source_windows as windows


class WindowTests(unittest.TestCase):
    def setUp(self):
        self.scope = {'side': 'SOURCE', 'execution_id': 'run_test',
                      'source_manifest_ref': 'sha256:' + 'a'*64,
                      'runtime_ref': 'sha256:' + 'b'*64, 'build_ref': 'sha256:' + 'c'*64}
        self.capture = windows.Windows(self.scope, lambda: {'side': 'SOURCE'}, 'sha256:' + 'd'*64)

    def valid(self, operation="same-step"):
        with patch.object(windows.time, 'monotonic_ns', side_effect=[0, 100, 1000100, 2000100, 3000100]), \
             patch.object(windows.time, 'time_ns', return_value=1700000000000000000):
            _, identity = self.capture.invoke(operation, lambda: {'status': 200})
        return json.loads(self.capture.records[identity])

    def test_anchor_rounding_and_no_later_wall_clock_read(self):
        record = self.valid()
        self.assertEqual(record['timer_elapsed_ms'], 3)
        self.assertEqual([e['elapsed_ms'] for e in record['events']], [2, 3])
        self.assertEqual(record['observation_window']['end'], '2023-11-14T22:13:20.003000Z')
        self.assertIsNone(record['observation_window']['schedule_horizon_ms'])

    def test_missing_reversed_forged_or_foreign_window_rejected(self):
        original = self.valid()
        mutations = {
            'start': lambda r: r['observation_window'].pop('start'),
            'end': lambda r: r['observation_window'].update(end=None),
            'reverse': lambda r: r.update(end_monotonic_ns=-1),
            'zero': lambda r: r.update(end_monotonic_ns=0),
            'execution': lambda r: r.update(execution_id='wrong'),
            'snapshot': lambda r: r.update(source_manifest_ref='wrong'),
            'runtime': lambda r: r.update(runtime_ref='wrong'),
            'build': lambda r: r.update(build_ref='wrong'),
            'outside': lambda r: r['events'][0].update(monotonic_ns=9000000),
            'incomplete': lambda r: r.update(post_capture=None),
            'target': lambda r: r.update(target_model_ref='poison'),
            'retrospective': lambda r: r.update(retrospective_start='invented'),
            'queue-fabrication': lambda r: r['observation_window'].update(due_beyond_horizon=False),
        }
        for name, mutate in mutations.items():
            with self.subTest(name=name):
                bad = copy.deepcopy(original)
                mutate(bad)
                with self.assertRaises((ValueError, KeyError, TypeError)):
                    windows.validate(bad, self.scope)

    def test_coherent_retiming_breaks_trusted_capture_hash(self):
        original = self.valid()
        bad = copy.deepcopy(original)
        bad['anchor']['utc_ns'] += 1000000000
        bad['observation_window']['start'] = windows.utc(bad['anchor']['utc_ns'])
        bad['observation_window']['end'] = windows.utc(bad['anchor']['utc_ns'] + bad['end_monotonic_ns'])
        with self.assertRaisesRegex(ValueError, 'artifact changed'):
            windows.validate(bad, self.scope, fixture.ref(fixture.canonical(original)))

    def test_repeated_visits_distinct_and_duplicate_payload_not_merged(self):
        a, b = self.valid(), self.valid()
        self.assertNotEqual(a['window_id'], b['window_id'])
        self.assertEqual([a['visit'], b['visit']], [1, 2])
        windows.validate_collection([b, a], self.scope)
        with self.assertRaises(ValueError):
            windows.validate_collection([a, a], self.scope)

    def test_reused_interval_cannot_move_to_another_execution(self):
        record = self.valid()
        with self.assertRaises(ValueError):
            windows.validate(record, {**self.scope, 'execution_id': 'different'})

    def test_failed_operation_retains_real_end_and_incomplete_status(self):
        with self.assertRaises(RuntimeError):
            self.capture.invoke('failure', lambda: (_ for _ in ()).throw(RuntimeError('source failure')))
        record = json.loads(next(iter(self.capture.records.values())))
        self.assertEqual(record['status'], 'INCOMPLETE')
        self.assertEqual(record['failure']['stage'], 'execution')
        self.assertIsNone(record['result'])
        self.assertIsNotNone(record['observation_window']['end'])
        with self.assertRaises(ValueError):
            windows.validate(record, self.scope)

    def test_capture_failure_cannot_be_empty_success(self):
        self.capture.capture = lambda: (_ for _ in ()).throw(RuntimeError('capture absent'))
        with self.assertRaises(RuntimeError):
            self.capture.invoke('failure', lambda: 0)
        record = json.loads(next(iter(self.capture.records.values())))
        self.assertIsNone(record['anchor'])
        self.assertEqual(record['status'], 'INCOMPLETE')

    def test_foreign_scope_rejected_before_operation(self):
        for name in ('target_model_ref', 'target_artifact_ref', 'comparator', 'plan_ref', 'approval_ref', 'timing_start'):
            with self.assertRaises(ValueError):
                windows.Windows({**self.scope, name: 'poison'}, lambda: {}, 'control')

    def test_concurrent_same_step_overlaps_without_identity_aliasing(self):
        barrier = __import__('threading').Barrier(2)
        with ThreadPoolExecutor(max_workers=2) as pool:
            futures = [pool.submit(self.capture.invoke, 'same', lambda: barrier.wait()) for _ in range(2)]
            for f in futures:
                f.result()
        records = [json.loads(raw) for raw in self.capture.records.values()]
        windows.validate_collection(records, self.scope)
        self.assertEqual({r['visit'] for r in records}, {1, 2})
        self.assertLess(max(r['anchor']['monotonic_ns'] for r in records), min(r['end_monotonic_ns'] for r in records))

    def test_wrong_operation_response_state_or_window_attachment_rejected(self):
        record = self.valid('HTTP /health')
        ref = fixture.ref(fixture.canonical(record))
        corpus = {'window_refs': [ref], 'observations': [{
            'window_ref': ref, 'window_id': record['window_id'],
            'request': {'path': '/health'}, 'response': record['result'],
            'before': record['pre_capture'], 'after': record['post_capture']}]}
        windows.validate_corpus(corpus, [record], self.scope)
        mutations = [lambda o: o.update(window_id='wrong'),
                     lambda o: o['request'].update(path='/another-actor'),
                     lambda o: o.update(response={'status': 403}),
                     lambda o: o.update(after={'unobserved': 'state'}),
                     lambda o: o.update(window_ref='sha256:' + '0'*64)]
        for mutate in mutations:
            bad = copy.deepcopy(corpus)
            mutate(bad['observations'][0])
            with self.assertRaises(ValueError):
                windows.validate_corpus(bad, [record], self.scope)

    def test_duplicate_physical_identity_refused(self):
        with patch.object(windows.uuid, 'uuid4') as identity:
            identity.return_value.hex = 'forced'
            self.valid()
            with self.assertRaises(ValueError):
                self.valid()


class FreshEvidenceTests(unittest.TestCase):
    def test_fresh_corpus_has_current_pins_and_distinct_integral_windows(self):
        root = fixture.FIXTURE / 'evidence/runs/windows-b1'
        objects = {fixture.ref(fixture.canonical(json.loads(p.read_bytes()))): json.loads(p.read_bytes())
                   for p in root.rglob('*.json')}
        q = json.loads((root / 'qualification.json').read_bytes())
        implementation = objects[q['implementation_ref']]
        self.assertEqual(implementation['source_windows.py'], fixture.ref(Path(windows.__file__).read_bytes()))
        result = objects[q['control_result_refs'][0]]
        self.assertTrue(result['reproduction_passed'])
        seen = set()
        projections = []
        for run in result['runs']:
            scope = {k: run[k] for k in ('side', 'execution_id', 'source_manifest_ref', 'runtime_ref', 'build_ref')}
            records = [objects[ref] for ref in run['window_refs']]
            corpus = objects[run['corpus_ref']]
            windows.validate_corpus(corpus, records, scope)
            self.assertEqual(len(records), run['window_count'])
            self.assertEqual(run['window_count'], 31)
            for record in records:
                self.assertNotIn(record['window_id'], seen)
                seen.add(record['window_id'])
                self.assertIsNone(record['observation_window']['schedule_horizon_ms'])
            self.assertTrue(all(c['ok'] for c in corpus['controls']))
            self.assertEqual(corpus['qualification_limits'], {'db_intermediate_visibility': 'PARTIAL', 'async_causal_completion_abi': 'PARTIAL'})
            failed = objects[run['failed_delivery_control_ref']]
            self.assertEqual(failed['status'], 'INCOMPLETE')
            self.assertIsNotNone(failed['observation_window']['end'])
            self.assertEqual(run['implementation_ref'], q['implementation_ref'])
            projections.append(corpus['reproduction_projection'])
        self.assertEqual(projections[0], projections[1])
        self.assertEqual(result['runs'][0]['build_ref'], result['runs'][1]['build_ref'])
        self.assertNotEqual(result['runs'][0]['execution_id'], result['runs'][1]['execution_id'])


if __name__ == '__main__':
    unittest.main()
