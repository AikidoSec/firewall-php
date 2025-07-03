--TEST--
Test \aikido\set_user

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php

\aikido\set_token("your token here")

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
