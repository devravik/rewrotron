"""Causal carrier qualification does not imply complete asynchronous observation."""
import copy
import unittest

import source_causal as causal


class CarrierTests(unittest.TestCase):
    def setUp(self):
        self.initial={'project_id':'ci3_reference','run_id':'run_one','workflow_instance_id':'run_one','operation_id':'create',
            'request_id':'request_one','process_id':'123','trace_id':'trace_one','span_id':'producer_one',
            'parent_span_id':None,'cause_ids':[],'job_id':None,'effect_id':None}
        self.queued={**self.initial,'job_id':'job_1','effect_id':'effect_1'}
        self.received={**self.queued,'process_id':'456','span_id':'worker_one','parent_span_id':'producer_one','cause_ids':['producer_one']}

    def test_valid_carrier(self):
        causal.validate_chain(self.initial,self.queued,self.received,'run_one','worker_one')

    def test_foreign_run_source_job_cause_and_span_refused(self):
        for field in ('project_id','run_id','trace_id','request_id','workflow_instance_id','operation_id','job_id','effect_id','parent_span_id','span_id','cause_ids'):
            changed=copy.deepcopy(self.received);changed[field]=['foreign'] if field=='cause_ids' else 'foreign'
            with self.assertRaises(ValueError):causal.validate_chain(self.initial,self.queued,changed,'run_one','worker_one')

    def test_missing_envelope_or_unrelated_queue_refused(self):
        changed=copy.deepcopy(self.queued);changed['span_id']='different'
        with self.assertRaises(ValueError):causal.validate_chain(self.initial,changed,self.received,'run_one','worker_one')
        with self.assertRaises(ValueError):causal.validate_chain(self.initial,{},self.received,'run_one','worker_one')

    def test_target_model_plan_expectations_cannot_be_carried(self):
        for field in ('target_model_ref','target_artifact_ref','plan_ref','approval_ref','expected_effects','comparator'):
            changed={**self.received,field:'poison'}
            with self.assertRaises(ValueError):causal.validate_chain(self.initial,self.queued,changed,'run_one','worker_one')

    def test_repeated_validation_is_readonly(self):
        original=copy.deepcopy((self.initial,self.queued,self.received))
        for _ in range(2):causal.validate_chain(self.initial,self.queued,self.received,'run_one','worker_one')
        self.assertEqual(original,(self.initial,self.queued,self.received))


class RuntimeCarrierTests(unittest.TestCase):
    def test_live_carriers_and_distinct_revision_are_hash_bound(self):
        from pathlib import Path
        import json
        import ci3_reference as fixture
        root=fixture.FIXTURE/'evidence/runs/causal-carrier-b1'
        objects={fixture.ref(fixture.canonical(json.loads(p.read_bytes()))):json.loads(p.read_bytes()) for p in root.rglob('*.json')}
        q=json.loads((root/'qualification.json').read_bytes())
        self.assertEqual(objects[q['implementation_ref']]['source_causal.py'],fixture.ref(Path(causal.__file__).read_bytes()))
        runs=objects[q['control_result_refs'][0]]['runs']
        for r in runs:
            causal.validate_chain(r['initiating'],r['queued'],r['received'],r['run_id'],r['worker_span'])
            self.assertNotEqual(r['source_manifest_ref'],r['original_source_ref'])
            self.assertEqual(r['completion'],'UNKNOWN')
            self.assertIsNone(r['horizon'])
            self.assertIsNone(r['due_beyond_horizon'])
            self.assertEqual(r['async_causal_completion_abi'],'PARTIAL')
            self.assertTrue(all(c['passed'] for c in r['role_controls']))
        self.assertEqual(runs[0]['build_ref'],runs[1]['build_ref'])
        self.assertNotEqual(runs[0]['run_id'],runs[1]['run_id'])
        with self.assertRaises(ValueError):
            causal.validate_chain(runs[0]['initiating'],runs[0]['queued'],runs[1]['received'],runs[0]['run_id'],runs[0]['worker_span'])


if __name__=='__main__':unittest.main()
