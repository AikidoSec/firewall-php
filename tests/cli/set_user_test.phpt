--TEST--
Test the aikido_set_user() function

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php

$_SERVER['REMOTE_ADDR'] = '::1';

$a = \aikido_set_user("122-sa-2", "username1");

if ($a == true) {
    echo "User set successfully\n";
} else {
    echo "User set failed\n";
}

?>

--EXPECTF--
[AIKIDO][INFO] [UEVENT] Got user event: 122-sa-2 username1 ::1
User set successfully