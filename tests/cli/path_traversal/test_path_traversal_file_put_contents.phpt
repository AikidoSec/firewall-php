--TEST--
Test path traversal (file_put_contents)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=../file

--FILE--
<?php

$file = '../file/test.txt';
    
file_put_contents($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
