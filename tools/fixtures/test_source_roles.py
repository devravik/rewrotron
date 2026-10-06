"""Falsification of the SOURCE capture handle qualification, not target execution."""
import copy
import json
from pathlib import Path
import unittest

import ci3_reference as fixture
import source_roles as roles


class RoleControlsTests(unittest.TestCase):
    def setUp(self):
        self.source_ref = 'sha256:' + 'a' * 64
        self.state = {'side': 'SOURCE', 'database': 'id\tname\n', 'source_files': {}, 'sink_files': {}}
        self.calls = 0

    def read(self):
        self.calls += 1
        return self.state

    def test_source_capture_executes_controls_without_mutating_state(self):
        before = copy.deepcopy(self.state)
        capture = roles.SourceCapture(self.read, self.source_ref)
        actual, results = roles.controls(capture, self.source_ref)
        self.assertGreater(self.calls, 1)
        self.assertTrue(all(r['passed'] for r in results))
        self.assertEqual(json.loads(actual), before)
        self.assertEqual(self.state, before)

    def test_contamination_rejected_before_observer_is_called(self):
        capture = roles.SourceCapture(self.read, self.source_ref)
        for key in roles.FORBIDDEN:
            with self.subTest(key=key):
                with self.assertRaises(ValueError):
                    capture.capture({'side': 'SOURCE', 'source_manifest_ref': self.source_ref, key: 'poison'})
        self.assertEqual(self.calls, 0)

    def test_plausible_capture_that_ignores_forbidden_fields_fails_qualification(self):
        read = self.read

        class IgnoresContamination:
            def capture(self, request):
                return fixture.canonical(read())

        with self.assertRaisesRegex(RuntimeError, 'reject-target_model_ref'):
            roles.controls(IgnoresContamination(), self.source_ref)

    def test_each_foreign_input_mutant_is_killed(self):
        for key in roles.FORBIDDEN:
            with self.subTest(key=key):
                good = roles.SourceCapture(self.read, self.source_ref)

                class AcceptsOneForeignInput:
                    def capture(self, request):
                        return good.capture({k: v for k, v in request.items() if k != key})

                with self.assertRaisesRegex(RuntimeError, 'reject-' + key):
                    roles.controls(AcceptsOneForeignInput(), self.source_ref)

    def test_foreign_provenance_is_rejected(self):
        for field in (*roles.FORBIDDEN, 'expected_state'):
            with self.subTest(field=field):
                observed = {**self.state, field: 'poison'}
                capture = roles.SourceCapture(lambda: observed, self.source_ref)
                with self.assertRaisesRegex(ValueError, 'provenance'):
                    capture.capture({'side': 'SOURCE', 'source_manifest_ref': self.source_ref})
        capture = roles.SourceCapture(lambda: {**self.state, 'side': 'TARGET'}, self.source_ref)
        with self.assertRaises(ValueError):
            roles.controls(capture, self.source_ref)

    def test_write_handle_mutant_is_killed(self):
        good = roles.SourceCapture(self.read, self.source_ref)

        class WritableCapture:
            def capture(self, request):
                return good.capture(request)

            def write(self, row):
                raise AssertionError('should never be invoked')

        with self.assertRaisesRegex(RuntimeError, 'observer-has-no-write'):
            roles.controls(WritableCapture(), self.source_ref)

    def test_missing_capture_never_qualifies(self):
        capture = roles.SourceCapture(lambda: {'side': 'SOURCE'}, self.source_ref)
        with self.assertRaisesRegex(ValueError, 'provenance'):
            roles.controls(capture, self.source_ref)

    def test_checker_read_is_detached_and_has_no_expected_fallback(self):
        actual = fixture.canonical(self.state)
        checker = roles.ActualReader(actual)
        check = checker.read()
        check['database'] = 'poison'
        self.assertEqual(fixture.canonical(checker.read()), actual)
        self.assertNotIn('principal', checker.read())

    def test_fixed_inputs_repeat_control_outcomes_and_capture_hash(self):
        a = roles.controls(roles.SourceCapture(self.read, self.source_ref), self.source_ref)
        b = roles.controls(roles.SourceCapture(self.read, self.source_ref), self.source_ref)
        self.assertEqual(a, b)
        self.assertEqual(fixture.ref(a[0]), fixture.ref(b[0]))

    def test_incomplete_duplicate_or_failed_control_results_cannot_qualify(self):
        good = [{'id': name, 'passed': True} for name in roles.required_controls()]
        roles.validate_results(good)
        for mutant in (good[:-1], good + [good[0]], [{'id': r['id'], 'passed': False} for r in good]):
            with self.assertRaises(ValueError):
                roles.validate_results(mutant)

    def test_qualification_hash_binds_changed_inputs(self):
        digest = lambda char: 'sha256:' + char * 64
        args = [digest('a'), [digest('b')], digest('c'), digest('d'), digest('e')]
        baseline = fixture.ref(fixture.canonical(roles.qualification(*args)))
        for i in range(len(args)):
            changed = copy.deepcopy(args)
            changed[i] = [digest('f')] if i == 1 else digest('f')
            self.assertNotEqual(baseline, fixture.ref(fixture.canonical(roles.qualification(*changed))))


class RuntimeEvidenceTests(unittest.TestCase):
    def test_retained_executions_bind_actual_artifacts_and_current_profile(self):
        results = []
        for run in range(2):
            root = fixture.FIXTURE / 'evidence/roles' / ('run-' + str(run))
            artifacts = {fixture.ref(fixture.canonical(json.loads(p.read_bytes()))): json.loads(p.read_bytes())
                         for p in root.glob('*.json')}
            index = json.loads((root / 'artifact-index.json').read_bytes())
            q = artifacts[index['qualification_ref']]
            result = artifacts[index['control_result_ref']]
            profile = artifacts[q['input_profile_ref']]
            implementation = artifacts[q['implementation_ref']]
            self.assertEqual(q['control_result_refs'], [index['control_result_ref']])
            self.assertEqual(q['capability_ids'], ['rwt-model-roles-1'])
            self.assertEqual(q['status'], 'SUPPORTED')
            self.assertEqual(result['implementation_ref'], q['implementation_ref'])
            self.assertEqual(result['control_suite_ref'], q['control_suite_ref'])
            self.assertEqual(result['profile_ref'], q['input_profile_ref'])
            self.assertEqual(profile['side'], 'SOURCE')
            self.assertEqual(profile['source_manifest_ref'], result['source_manifest_ref'])
            self.assertEqual(profile['runtime_ref'], fixture.ref((fixture.FIXTURE / 'runtime.lock.json').read_bytes()))
            self.assertEqual(implementation['source_roles.py'], fixture.ref(Path(roles.__file__).read_bytes()))
            self.assertEqual(implementation['ci3_reference.py'], fixture.ref(Path(fixture.__file__).read_bytes()))
            self.assertEqual(artifacts[result['source_manifest_ref']], fixture.source_manifest())
            self.assertEqual(artifacts[result['launch_ref']]['Image'], result['loaded_image_ref'])
            self.assertIn('/usr/local/bin/php', artifacts[result['build_ref']])
            self.assertEqual(artifacts[result['capture_ref']]['side'], 'SOURCE')
            self.assertEqual(result['db_write_refusal']['stderr_ref'], fixture.ref((root / 'db-write-refusal.stderr').read_bytes()))
            self.assertEqual(result['observer_server_audit_ref'], fixture.ref((root / 'observer-server-audit.tsv').read_bytes()))
            roles.validate_results(result['controls'])
            self.assertEqual(artifacts[q['control_suite_ref']]['required_controls'],
                             [{'id': name, 'passed': True} for name in roles.required_controls()])
            results.append(result)
        for key in ('implementation_ref', 'profile_ref', 'source_manifest_ref', 'runtime_ref', 'build_ref', 'capture_ref', 'controls', 'runtime'):
            self.assertEqual(results[0][key], results[1][key], key)
        self.assertNotEqual(results[0]['execution_id'], results[1]['execution_id'])
        self.assertNotEqual(results[0]['launch_ref'], results[1]['launch_ref'])


if __name__ == '__main__':
    unittest.main()
