--TEST--
Test path traversal 2

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCKING=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = '../file';
$file1 = 'aa.txt';
$file2 = '../file/test.txt';

copy($file1, $file2);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Path traversal detected.*

