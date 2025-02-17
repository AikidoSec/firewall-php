<?php

$_SERVER['REMOTE_ADDR'] = '5.2.190.71';

if (extension_loaded('aikido')) {
    \aikido\set_user("12345", "Tudor");

    $decision = \aikido\should_block_request();

    // If the rate limit is exceeded, return a 429 status code
    if ($decision->block && $decision->type == "ratelimited" && $decision->trigger == "user") {
        http_response_code(429);
        echo "Rate limit exceeded";
        exit();
    }
}

// Continue handling the request
echo "Request successful!";

?>
