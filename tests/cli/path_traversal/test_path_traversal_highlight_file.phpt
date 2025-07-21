--TEST--
Test path traversal (highlight_file)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=../file/test.txt

--FILE--
<?php

$file = '../file/test.txt';
    
highlight_file($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
