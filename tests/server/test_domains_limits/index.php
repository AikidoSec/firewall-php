<?php

$ch1 = curl_init("http://" . $_GET['domain'] . ".com/");
curl_setopt($ch1, CURLOPT_RETURNTRANSFER, false);
curl_setopt($ch1, CURLOPT_HEADER, false);
curl_setopt($ch1, CURLOPT_TIMEOUT_MS, 1);
curl_setopt($ch1, CURLOPT_CONNECTTIMEOUT_MS, 1);
curl_exec($ch1);
curl_close($ch1);

?>
