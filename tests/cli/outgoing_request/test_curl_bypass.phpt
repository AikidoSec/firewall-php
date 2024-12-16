--TEST--
Test curl exec function

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php

$ch1 = curl_init("http://localhost:1337/?id[]=google.com;curl%20http://google.com");
curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
curl_exec($ch1);
curl_close($ch1);

?>

--EXPECT--
[AIKIDO][INFO] [BEFORE] Got domain: example.com
[AIKIDO][INFO] [AFTER] Got domain: example.com port: 443
