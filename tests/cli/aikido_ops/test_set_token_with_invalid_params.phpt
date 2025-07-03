--TEST--
Test \aikido\set_token without parameter

--ENV--
AIKIDO_LOG_LEVEL=INFO

--SKIPIF--
<?php if (PHP_VERSION_ID < 70400) die("PHP >= 7.4."); ?>

--FILE--
<?php

\aikido\set_token();

?>

--EXPECTREGEX--
.*Fatal error: Uncaught ArgumentCountError: aikido\\set_token\(\) expects exactly 1 argument, 0 given.*