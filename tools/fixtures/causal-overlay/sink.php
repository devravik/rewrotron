<?php
// Closed disposable receiver. Never delegates to a real email/payment endpoint.
$p = json_decode(file_get_contents('php://input'), true);
if (!is_array($p) || !isset($p['business_key'], $p['item'])) { http_response_code(400); exit; }
$capture = json_decode(base64_decode(isset($_SERVER['HTTP_X_REWROTRON_CAPTURE']) ? $_SERVER['HTTP_X_REWROTRON_CAPTURE'] : ''), true);
if (!is_array($capture)) { http_response_code(400); exit; }
$key = $p['business_key']; $id = hash('sha256', $key);
$file = '/state/sink/'.$id.'.json';
$applied = !is_file($file);
if ($applied) { file_put_contents($file, json_encode(['payload' => $p, 'business_application_id' => $id, 'capture_context' => $capture]), LOCK_EX); }
header('Content-Type: application/json');
echo json_encode(['receipt' => $id, 'business_application_id' => $id, 'applied' => $applied]);
