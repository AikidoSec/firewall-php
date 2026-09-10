<?php

if (empty($_SERVER['FRANKENPHP_WORKER'])) {
    http_response_code(200);
    echo json_encode(['skip' => 'Requires FrankenPHP worker mode']);
    return;
}

// Globals persist between requests handled by the same worker.
global $workerId;
$workerId = $workerId ?? bin2hex(random_bytes(8));

$blocked = false;
try {
    file_get_contents(__DIR__ . '/' . $_GET['path']);
} catch (Throwable $e) {
    if (strpos($e->getMessage(), 'Aikido firewall has blocked a path traversal attack') === false) {
        throw $e;
    }
    $blocked = true;
}
http_response_code(200);
header('Content-Type: application/json');
echo json_encode(['worker' => $workerId, 'blocked' => $blocked]);
