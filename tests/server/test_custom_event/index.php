<?php

\aikido\track("user.login_failed");
\aikido\set_user("user-1", "Jane Doe");
\aikido\track("user.login_succeeded");
\aikido\track("");

$decision = \aikido\should_block_request();
if ($decision->block && $decision->type == "blocked" && $decision->trigger == "user") {
    http_response_code(403);
    echo "User blocked after events were tracked\n";
    exit();
}

echo "User was not blocked\n";
