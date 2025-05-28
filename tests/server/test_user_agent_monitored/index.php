<?php

if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    if ($decision->block && $decision->type == "blocked" && $decision->trigger == "user-agent") {
        http_response_code(403);
        echo "Your user agent ({$decision->user_agent}) is blocked due to: {$decision->description}!";
        exit();
    }
}

echo "Something";

?>
