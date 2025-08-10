<?php

$_SERVER['REMOTE_ADDR'] = sprintf(
    '%d.%d.%d.%d',
    rand(1, 255),
    rand(0, 255),
    rand(0, 255),
    rand(1, 254)
);

\aikido\set_rate_limit_group("my_user_group");

if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    // If the rate limit is exceeded, return a 429 status code
    if ($decision->block && $decision->type == "ratelimited" && $decision->trigger == "group") {
        http_response_code(429);
        echo "Rate limit exceeded";
        exit();
    }
}

// Continue handling the request
echo "Request successful!";

?>