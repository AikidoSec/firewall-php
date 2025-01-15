--TEST--
Test PHP shell injection (popen)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=www.example`whoami`.com

--FILE--
<?php

$handle = popen('binary --domain www.example`whoami`.com', 'r');
while (!feof($handle)) {
    $buffer = fgets($handle);
    echo $buffer;
}
pclose($handle);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a shell injection.*