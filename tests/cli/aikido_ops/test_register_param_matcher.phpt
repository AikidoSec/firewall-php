--TEST--
Test \aikido\register_param_matcher with valid parameters

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_TOKEN=AIK_DUMMY

--FILE--
<?php

$result = \aikido\register_param_matcher("param_name", "{digits}-{alpha}");

?>

--EXPECTREGEX--
\[AIKIDO\]\[INFO\]\[tid:\d+\] Token changed to "AIK_RUNTIME_\*\*\*UMMY"
\[AIKIDO\]\[INFO\]\[tid:\d+\] Registered param matcher param_name -> \{digits\}-\{alpha\}