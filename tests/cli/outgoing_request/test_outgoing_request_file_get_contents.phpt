--TEST--
Test outgoing request (file_get_contents)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php
    
file_get_contents("http://www.example.com");

?>

--EXPECTREGEX--
.*\[AIKIDO\]\[INFO\]\[tid:\d+\] \[BEFORE\] Got domain: www.example.com
\[AIKIDO\]\[INFO\]\[tid:\d+\] \[AFTER\] Got domain: www.example.com port: 80