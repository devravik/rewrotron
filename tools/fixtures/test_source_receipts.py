"""Receiver evidence is created only by actual independent HTTP delivery."""
import base64
import copy
import json
from pathlib import Path
import tempfile
import time
import unittest
from urllib import request,error

import ci3_reference as fixture
import source_receipts as receipts


class ReceiptTests(unittest.TestCase):
    def setUp(self):
        self.temporary=tempfile.TemporaryDirectory();self.addCleanup(self.temporary.cleanup)
        self.scope={'side':'SOURCE','execution_id':'run_one','window_id':'window_one','source_manifest_ref':'sha256:'+'a'*64,'runtime_ref':'sha256:'+'b'*64,'build_ref':'sha256:'+'c'*64}
        self.initial={'project_id':'ci3_reference','run_id':'run_one','workflow_instance_id':'run_one','operation_id':'create','request_id':'request_one',
            'process_id':'123','trace_id':'trace_one','span_id':'create_one','parent_span_id':None,'cause_ids':[],'job_id':None,'effect_id':None}
        self.queued={**self.initial,'job_id':'job_1','effect_id':'effect_1'}
        self.received={**self.queued,'process_id':'456','span_id':'worker_one','parent_span_id':'create_one','cause_ids':['create_one']}
        self.anchor={'monotonic_ns':time.monotonic_ns(),'utc_ns':time.time_ns()}
        self.receiver=receipts.Receiver('127.0.0.1',Path(self.temporary.name)/'receiver',self.scope,self.initial,'worker_one',self.anchor)
        self.receiver.bind_queue({'id':1,'item':1,'capture_context':json.dumps(self.queued)})
        self.receiver.__enter__();self.addCleanup(self.receiver.__exit__)

    def send(self,envelope=None,payload=None):
        body=fixture.canonical(payload or {'business_key':'item-1','item':1})
        headers={'X-Rewrotron-Capture':base64.b64encode(fixture.canonical(envelope or self.received)).decode()}
        return request.urlopen(request.Request('http://127.0.0.1:'+str(self.receiver.server.server_port)+'/',data=body,headers=headers),timeout=3)

    def events(self):
        # Handler appends send event after flushing response; join via its lock/short bounded polling.
        for _ in range(100):
            with self.receiver.lock:
                if len(self.receiver.events)==2:return copy.deepcopy(self.receiver.events)
            time.sleep(.001)
        self.fail('ACK boundary not recorded')

    def test_actual_receipt_and_distinct_ack_response_send(self):
        with self.send() as response:ack=json.load(response)
        events=self.events()
        receipts.validate_events(events,self.scope,self.anchor,fixture.ref(fixture.canonical(events)))
        self.assertEqual(ack['receipt'],events[0]['receipt_id'])
        self.assertEqual(ack['ack_id'],events[1]['ack_id'])
        self.assertGreaterEqual(events[1]['sent_monotonic_ns'],events[0]['received_monotonic_ns'])
        self.assertEqual(len(list(self.receiver.storage.iterdir())),1)
        self.assertNotIn('completion',ack)

    def test_http_success_before_worker_is_not_receiver_receipt(self):
        self.assertEqual(self.receiver.events,[])
        self.assertEqual(list(self.receiver.storage.iterdir()),[])
        with self.assertRaises(ValueError):receipts.validate_events([],self.scope,self.anchor,fixture.ref(b'[]'))

    def test_wrong_job_run_request_cause_span_and_foreign_fields_rejected(self):
        for key in ('job_id','run_id','request_id','parent_span_id','span_id','target_model_ref','plan_ref','approval_ref','expected_ack'):
            bad={**self.received,key:'foreign'}
            with self.assertRaises(error.HTTPError):self.send(bad)
            self.assertEqual(self.receiver.events,[])
        with self.assertRaises(error.HTTPError):self.send(payload={'business_key':'item-2','item':2})

    def test_replay_cannot_create_second_application_or_ack(self):
        with self.send() as response:response.read()
        first=self.events()
        with self.assertRaises(error.HTTPError):self.send()
        self.assertEqual(first,self.receiver.events)

    def test_altered_timestamps_scope_ack_and_defaults_break_hash(self):
        with self.send() as response:response.read()
        actual=self.events();ref=fixture.ref(fixture.canonical(actual))
        for key in ('received_monotonic_ns','execution_id','window_id','source_manifest_ref','runtime_ref','build_ref','receipt_id'):
            bad=copy.deepcopy(actual);bad[0][key]='altered'
            with self.assertRaises(ValueError):receipts.validate_events(bad,self.scope,self.anchor,ref)
        bad=copy.deepcopy(actual);bad[1]['receipt_id']='wrong'
        with self.assertRaises(ValueError):receipts.validate_events(bad,self.scope,self.anchor,fixture.ref(fixture.canonical(bad)))

    def test_missing_ack_not_completion(self):
        with self.send() as response:response.read()
        first=self.events()[:1]
        with self.assertRaises(ValueError):receipts.validate_events(first,self.scope,self.anchor,fixture.ref(fixture.canonical(first)))


class RuntimeReceiptTests(unittest.TestCase):
    def test_runtime_event_hashes_durable_storage_and_cross_run_binding(self):
        root=fixture.FIXTURE/'evidence/runs/receipt-ack-b1'
        objects={fixture.ref(fixture.canonical(json.loads(p.read_bytes()))):json.loads(p.read_bytes()) for p in root.rglob('*.json')}
        q=json.loads((root/'qualification.json').read_bytes())
        self.assertEqual(objects[q['implementation_ref']]['source_receipts.py'],fixture.ref(Path(receipts.__file__).read_bytes()))
        runs=objects[q['control_result_refs'][0]]['runs']
        for run in runs:
            scope={k:run[k] for k in ('side','execution_id','window_id','source_manifest_ref','runtime_ref','build_ref')}
            events=objects[run['receiver_events_ref']]
            receipts.validate_events(events,scope,run['anchor'],run['receiver_events_ref'])
            self.assertEqual(run['producer_forged_ack_http_status'],409)
            file=next((root/run['execution_id']/'receiver-store').iterdir())
            self.assertEqual(fixture.ref(file.read_bytes()),events[0]['durable_artifact_ref'])
            self.assertEqual(run['completion'],'UNKNOWN')
            for field in ('horizon','quiescence','due_beyond_horizon'):self.assertIsNone(run[field])
        self.assertEqual(runs[0]['build_ref'],runs[1]['build_ref'])
        self.assertNotEqual(runs[0]['execution_id'],runs[1]['execution_id'])
        scope={k:runs[1][k] for k in ('side','execution_id','window_id','source_manifest_ref','runtime_ref','build_ref')}
        with self.assertRaises(ValueError):receipts.validate_events(objects[runs[0]['receiver_events_ref']],scope,runs[1]['anchor'],runs[0]['receiver_events_ref'])


if __name__=='__main__':unittest.main()
