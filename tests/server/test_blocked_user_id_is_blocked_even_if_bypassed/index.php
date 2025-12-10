<?php

if (extension_loaded('aikido')) {
    \aikido\set_user("12345", "Tudor");

    $decision = \aikido\should_block_request();

    // If the user is blocked, return a 403 status code
    if ($decision->block && $decision->type == "blocked" && $decision->trigger == "user") {
        http_response_code(403);
        echo "You are blocked by Aikido Firewall!";
        exit();
    }
}

// Continue handling the request
echo "User set successfully";
?>
