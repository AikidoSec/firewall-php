<?php

if (extension_loaded('aikido')) {
    $decision = \aikido\should_whitelist_request();

    if ($decision == null) {
        echo "Decision is null!";
    } else {
        echo "whitelisted=" . ($decision->whitelisted ? "true" : "false") . ";";
        echo "type=" . $decision->type . ";";
        echo "trigger=" . $decision->trigger . ";";
        echo "description=" . $decision->description . ";";
        echo "ip=" . $decision->ip . ";";
    }
}

echo "Something!";

?>
