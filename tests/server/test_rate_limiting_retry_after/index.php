<?php

$_SERVER['REMOTE_ADDR'] = $_SERVER['HTTP_X_FORWARDED_FOR'] ?? '4.18.92.7';

if (isset($_GET['user'])) {
    \aikido\set_user($_GET['user']);
}
if (isset($_GET['group'])) {
    \aikido\set_rate_limit_group($_GET['group']);
}

$decision = \aikido\should_block_request();
if ($decision->block) {
    if ($decision->type === 'ratelimited') {
        http_response_code(429);
        if (($decision->retry_after ?? 0) > 0) {
            header('Retry-After: ' . $decision->retry_after);
        }
    } else {
        http_response_code(403);
    }
}
header('Content-Type: application/json');
echo json_encode($decision);
