--TEST--
Test path traversal (symlink)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=../file

--FILE--
<?php

$file = 'test.txt';
$dest = '../file/test.txt';
    
symlink($file, $dest);
    

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
