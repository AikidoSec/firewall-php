--TEST--
Test PHP shell injection (exec)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

$_SERVER['HTTP_USER'] = 'www.example`whoami`.com';

$output = array();
$return_var = 0;
exec('binary --domain www.example`whoami`.com', $output, $return_var);
print_r($output);
echo "\n";

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a shell injection.*