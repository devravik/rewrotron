<?php
// Application worker; only this code delivers jobs. Observer never inserts effects.
$pdo = new PDO('mysql:host=db;dbname=fixture;charset=utf8mb4', 'fixture', 'synthetic-fixture-only');
foreach ($pdo->query("SELECT id,item FROM jobs WHERE status='pending' ORDER BY id") as $job) {
    $payload = json_encode(['business_key' => 'item-'.$job['item'], 'item' => (int)$job['item']]);
    $c = stream_context_create(['http' => ['method' => 'POST', 'header' => 'Content-Type: application/json', 'content' => $payload, 'timeout' => 2]]);
    $receipt = file_get_contents('http://sink:8081/', false, $c);
    if ($receipt === false) { exit(1); }
    $ack = json_decode($receipt, true);
    $q = $pdo->prepare("UPDATE jobs SET status='acked',receipt=? WHERE id=?");
    $q->execute([$ack['receipt'], $job['id']]);
}
