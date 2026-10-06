<?php
// Application worker; only this code delivers jobs. Observer never inserts effects.
$pdo = new PDO('mysql:host=db;dbname=fixture;charset=utf8mb4', 'fixture', 'synthetic-fixture-only');
foreach ($pdo->query("SELECT id,item,capture_context FROM jobs WHERE status='pending' ORDER BY id") as $job) {
    $payload = json_encode(['business_key' => 'item-'.$job['item'], 'item' => (int)$job['item']]);
    $capture = json_decode($job['capture_context'], true);
    if (!is_array($capture)) { exit(2); }
    $capture['parent_span_id'] = $capture['span_id'];
    $capture['cause_ids'] = [$capture['span_id']];
    $capture['span_id'] = getenv('RWT_WORKER_SPAN');
    $capture['process_id'] = (string)getmypid();
    $header = 'Content-Type: application/json'."\r\n".'X-Rewrotron-Capture: '.base64_encode(json_encode($capture));
    $c = stream_context_create(['http' => ['method' => 'POST', 'header' => $header, 'content' => $payload, 'timeout' => 2]]);
    $receipt = file_get_contents('http://sink:8081/', false, $c);
    if ($receipt === false) { exit(1); }
    $ack = json_decode($receipt, true);
    $q = $pdo->prepare("UPDATE jobs SET status='acked',receipt=? WHERE id=?");
    $q->execute([$ack['receipt'], $job['id']]);
}
