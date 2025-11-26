--TEST--
Test \aikido\register_param_matcher with valid parameters

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_TOKEN=AIK_DUMMY

--FILE--
<?php

$result = \aikido\register_param_matcher("param_name", "{digits}-{alpha}");

?>

--EXPECT--
[AIKIDO][INFO] Token changed to "AIK_RUNTIME_***UMMY"
Error compiling param matcher no_braces -> regex "digits-alpha": pattern should contain { or }
bool(false)
Error compiling param matcher unclosed_brace -> regex "{digits": pattern should contain { or }
bool(false)
Error compiling param matcher with_slash -> regex "aikido/{digits}": pattern should not contain slashes
bool(false)