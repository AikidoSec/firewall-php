<?php

$_SERVER['HTTP_USER_AGENT'] = "1234googlebot1234";

if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    if ($decision->block && $decision->type == "blocked" && $decision->trigger == "user-agent") {
        http_response_code(403);
        echo "Your user agent ({$decision->data}) is blocked due to: {$decision->description}!";
        exit();
    }
}

echo "Something";

?>
