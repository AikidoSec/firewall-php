--TEST--
Test ssrf (file_get_contents with multiple php filter)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=https://localhost:8081

--FILE--
<?php

$file = 'php://filter/convert.base64-encode/resource=php://filter/resource=php://filter/resource=php://filter/resource=https://localhost:8081';
file_get_contents($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a server-side request forgery.*
