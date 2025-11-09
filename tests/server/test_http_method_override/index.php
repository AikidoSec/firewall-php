<?php


$method = $_SERVER['REQUEST_METHOD'];
$_SERVER['REMOTE_ADDR'] = '5.2.190.71';

if ($method === 'POST') {
    if (isset($_GET['_method'])) {
        $method = strtoupper($_GET['_method']);
    }

    else if (isset($_SERVER['HTTP_X_HTTP_METHOD_OVERRIDE'])) {
        $method = strtoupper($_SERVER['HTTP_X_HTTP_METHOD_OVERRIDE']);
    }
}


if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    // If the rate limit is exceeded, return a 429 status code
    if ($decision->block && $decision->type == "ratelimited" && $decision->trigger == "ip") {
        http_response_code(429);
        echo "Rate limit exceeded by IP: " . $decision->ip;
        exit();
    }
}

echo "Method: " . $method;

