<?php

if (extension_loaded('aikido')) {
    $decision = \aikido\should_block_request();

    if ($decision == null) {
        echo "Decision is null!";
    }
}

echo "Something!";

?>
