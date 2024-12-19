--TEST--
Test PHP shell injection (proc_open)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=www.example`whoami`.com

--FILE--
<?php

$descriptorspec = array(
    0 => array("pipe", "r"),  // stdin is a pipe that the child will read from
    1 => array("pipe", "w"),  // stdout is a pipe that the child will write to
    2 => array("pipe", "w")   // stderr is a pipe that the child will write to
);

$process = proc_open('binary --domain www.example`whoami`.com', $descriptorspec, $pipes);

if (is_resource($process)) {
    while ($s = fgets($pipes[1])) {
        echo $s;
    }
    fclose($pipes[1]);
    proc_close($process);
}

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a shell injection.*