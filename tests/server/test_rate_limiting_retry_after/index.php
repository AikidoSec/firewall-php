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
    http_response_code($decision->type === 'ratelimited' ? 429 : 403);
    if ($decision->retry_after !== null) {
        header('Retry-After: ' . $decision->retry_after);
    }
}
header('Content-Type: application/json');
echo json_encode($decision);
