<?php

for ($i = 0; $i < 10; $i++) {
    $ch1 = curl_init("https://example.com/");
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch1, CURLOPT_TIMEOUT, 5);
    $response = curl_exec($ch1);
    curl_close($ch1);
}

$ch2 = curl_init("https://ssrf-redirects.testssandbox.com/ssrf-test-4");
curl_setopt($ch2, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch2, CURLOPT_TIMEOUT, 5);
curl_setopt($ch2, CURLOPT_FOLLOWLOCATION, true);
curl_exec($ch2);
curl_close($ch2);

$ch3 = curl_init();
$options = [
CURLOPT_URL => "https://facebook.com",
CURLOPT_RETURNTRANSFER => true,
CURLOPT_TIMEOUT => 5,
CURLOPT_HEADER => false,
];
curl_setopt_array($ch3, $options);
curl_exec($ch3);
curl_close($ch3);

$ch4 = curl_init("https://facebook.com:443");
curl_setopt($ch4, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch4, CURLOPT_TIMEOUT, 5);
curl_exec($ch4);
curl_close($ch4);

$ch5 = curl_init("http://www.aikido.dev:80");
curl_setopt($ch5, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch5, CURLOPT_TIMEOUT, 5);
curl_exec($ch5);
curl_close($ch5);

$ch6 = curl_init("http://some-invalid-domain.com:4113");
curl_setopt($ch6, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch6, CURLOPT_TIMEOUT, 5);
curl_exec($ch6);
curl_close($ch6);

for ($i = 0; $i < 3; $i++) {
    $ch5 = curl_init("http://www.aikido.dev:443");
    curl_setopt($ch5, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch5, CURLOPT_TIMEOUT, 5);
    curl_exec($ch5);
    curl_close($ch5);
}



?>
