--TEST--
Test path traversal (include_once)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = '../file';

$file = '../file/test.txt';
    
include_once($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
