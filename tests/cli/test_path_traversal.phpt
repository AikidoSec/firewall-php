--TEST--
Test path traversal

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCKING=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = '../file';

$file = '../file/test.txt';
chmod($file, 0777);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Path traversal detected.*

