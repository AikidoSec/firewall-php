--TEST--
Test PHP shell execution functions (json decode)

--POST_RAW--
{
    "age": -123123e10000,
    "cmd": "cat /etc/passwd"
}

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1 


--FILE--
<?php
$input = file_get_contents('php://input');
$cmd = json_decode($input, true)['cmd'];

echo exec($cmd) . "\n";

?>

--EXPECTREGEX--
.*Aikido firewall has blocked a shell injection.*