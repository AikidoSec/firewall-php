--TEST--
Test path traversal (file_get_contents with php filter)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

$file = '../file/test.txt';
    
file_get_contents("php://filter/resource=" . $file);
//file_get_contents("php://filter/resource=php://filter/resource=php://filter/resource=php://filter/resource=php://filter/resource=php://filter/resource=php://filter/resource=../test.txt");

?>

--POST--
test=../file

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a path traversal attack.*
