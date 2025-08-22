--TEST--
Test path ssrf \@  in authority

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=http://127.1.1.1:4000\@127.0.0.1:80/

--FILE--
<?php

$file = 'http://127.1.1.1:4000\@127.0.0.1:80/';
    
file_get_contents($file);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a server-side request forgery.*
