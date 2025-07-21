--TEST--
Test path traversal (SplFileObject)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

$file = '../file/test.txt';

new SplFileObject($file, 'r', false, null);

?>

--POST--
test=../file

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
