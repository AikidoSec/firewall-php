--TEST--
Disabling Aikido does not log a request processor initialization error

--ENV--
AIKIDO_DISABLE=true
AIKIDO_DEBUG=0
AIKIDO_LOG_LEVEL=ERROR
AIKIDO_DISK_LOGS=0

--FILE--
<?php
echo "Request completed\n";
?>

--EXPECT--
Request completed
