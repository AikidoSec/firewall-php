--TEST--
Test PHP shell injection (shell_exec)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = 'www.example`whoami`.com';

$output = shell_exec('binary --domain www.example`whoami`.com');
echo $output;

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a shell injection.*