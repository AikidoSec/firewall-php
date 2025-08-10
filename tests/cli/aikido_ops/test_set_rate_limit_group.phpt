--TEST--
Test \aikido\set_rate_limit_group

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php

\aikido\set_rate_limit_group("my_user_group")

?>

--EXPECT--
[AIKIDO][INFO] Got rate limit group: my_user_group
