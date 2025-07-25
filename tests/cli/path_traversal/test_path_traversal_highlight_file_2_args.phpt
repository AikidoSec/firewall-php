--TEST--
Test path traversal (highlight_file with two args)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=../file/test.txt

--FILE--
<?php

$file = '../file/test.txt';
    
highlight_file($file, false);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
