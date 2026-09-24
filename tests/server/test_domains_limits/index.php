<?php

$ch1 = curl_init("http://" . $_GET['domain'] . ".com/");
curl_setopt($ch1, CURLOPT_RETURNTRANSFER, false);
curl_setopt($ch1, CURLOPT_HEADER, false);
// Test hostname retention without waiting for public DNS or remote servers.
curl_setopt($ch1, CURLOPT_CONNECT_TO, ["::127.0.0.1:1"]);
curl_setopt($ch1, CURLOPT_PROXY, "");
curl_setopt($ch1, CURLOPT_TIMEOUT_MS, 1);
curl_setopt($ch1, CURLOPT_CONNECTTIMEOUT_MS, 1);
curl_exec($ch1);
curl_close($ch1);

?>
