--TEST--
Test \aikido\register_param_matcher with valid parameters

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php

$result = \aikido\register_param_matcher("param_name", "{digits}-{alpha}");

?>

--EXPECT--
[AIKIDO][INFO] Got param matcher name: param_name
[AIKIDO][INFO] Got param matcher regex: {digits}-{alpha}
