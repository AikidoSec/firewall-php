--TEST--
Test path traversal (chgrp)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCKING=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = '../file';

$file = '../file/test.txt';
    
chgrp($file, 0);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Path traversal detected.*