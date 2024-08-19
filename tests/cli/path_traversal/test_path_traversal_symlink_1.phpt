--TEST--
Test path traversal (symlink)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCKING=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = '../file';

$file = '../file/test.txt';
$dest = 'test2.txt';
    
symlink($file, $dest);
    

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Path traversal detected.*
