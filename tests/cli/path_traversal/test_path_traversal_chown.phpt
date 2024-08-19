--TEST--
Test path traversal (chown)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCKING=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = '../file';

$file = '../file/test.txt';
    
chown($file, 0);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Path traversal detected.*