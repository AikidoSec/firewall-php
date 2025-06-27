--TEST--
Test path ssrf (file_get_contents) - Case Insensitive

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=hTTps://lOcalhosT:8081

--FILE--
<?php

$file = 'hTTps://lOcalhosT:8081';
    
file_get_contents($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a server-side request forgery.*
