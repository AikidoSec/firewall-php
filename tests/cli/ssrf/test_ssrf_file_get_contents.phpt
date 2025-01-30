--TEST--
Test path ssrf (file_get_contents)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=http://127.0.0.1:8081

--FILE--
<?php

$file = 'http://127.0.0.1:8081';
    
file_get_contents($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a server-side request forgery.*
