<?php

$_SERVER['REMOTE_ADDR'] = '5.2.190.71';

if (extension_loaded('aikido')) {
    \aikido\set_user("12345", "Tudor");

    $decision = \aikido\should_block_request();

    if ($decision->block && $decision->type == "ratelimited") {
        http_response_code(429);
        header("Retry-After: " . $decision->retry_after_seconds);
        echo "Rate limit exceeded";
        exit();
    }
}

echo "Request successful!";

?>
