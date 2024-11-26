--TEST--
Test PHP shell execution functions

--ENV--
AIKIDO_LOG_LEVEL=INFO

--SKIPIF--
<?php if (PHP_VERSION_ID < 70400) die("skip Array can be passed as parameter to proc_open instead of string only from PHP >= 7.4."); ?>

--FILE--
<?php
// proc_open()
$descriptorspec = array(
    0 => array("pipe", "r"),  // stdin is a pipe that the child will read from
    1 => array("pipe", "w"),  // stdout is a pipe that the child will write to
    2 => array("pipe", "w")   // stderr is a pipe that the child will write to
);

$command = ["echo", "test"];
$process = proc_open($command, $descriptorspec, $pipes);

?>

--EXPECTF--
