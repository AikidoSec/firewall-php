<?php

$_SERVER['REMOTE_ADDR'] = '192.42.116.197';

if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    if ($decision->block && $decision->type == "blocked") {
        http_response_code(403);
        echo "Your IP ({$decision->ip}) is blocked due to: {$decision->description}!";
        exit();
    }
}

echo "Something";

?>
