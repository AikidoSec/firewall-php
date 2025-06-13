<?php

if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    // If the rate limit is exceeded, return a 429 status code
    if ($decision->block && $decision->type == "ratelimited" && $decision->trigger == "ip") {
        http_response_code(429);
        echo "Rate limit exceeded";
        exit();
    }
}

// Continue handling the request
echo "Request successful";

?>
