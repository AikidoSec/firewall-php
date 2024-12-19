--TEST--
Test path traversal (rename)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=../file

--FILE--
<?php

$file = 'test.txt';
$dest = '../file/test.txt';
    
rename($file, $dest);
    

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
