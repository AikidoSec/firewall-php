--TEST--
Test path traversal (php_strip_whitespaces)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

$file = '../file/test.txt';

php_strip_whitespace($file);

?>

--POST--
test=../file/test.txt

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
