--TEST--
Test path ssrf (curl_exec)

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=http://127.0.0.1:8081

--FILE--
<?php
    
$ch1 = curl_init("http://127.0.0.1:8081");
curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
curl_exec($ch1);
curl_close($ch1);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a server-side request forgery.*
