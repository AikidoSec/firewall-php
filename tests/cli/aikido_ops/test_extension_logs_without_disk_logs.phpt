--TEST--
PHP extension logs errors when disk logs are disabled

--ENV--
AIKIDO_LOG_LEVEL=ERROR
AIKIDO_DISK_LOGS=0

--FILE--
<?php

\aikido\set_token("");

?>

--EXPECTREGEX--
.*\[AIKIDO\]\[ERROR\]\[\d+\]\[\d+\]\[[0-9:]+\] set_token: token is null!.*
