--TEST--
Test curl exec function

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php
$ch1 = curl_init("https://example.com/");
curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch1, CURLOPT_TIMEOUT, 1);
curl_exec($ch1);
curl_close($ch1);


$ch2 = curl_init("https://httpbin.org/get");
$queryParams = http_build_query([
    'param1' => 'value1',
    'param2' => 'value2'
]);
curl_setopt($ch2, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch2, CURLOPT_URL, "https://httpbin.org/get?" . $queryParams);
curl_setopt($ch2, CURLOPT_TIMEOUT, 1);
curl_exec($ch2);
curl_close($ch2);

$ch3 = curl_init();
$options = [
CURLOPT_URL => "https://facebook.com",
CURLOPT_RETURNTRANSFER => true,
CURLOPT_HEADER => false,
CURLOPT_TIMEOUT => 1
];
curl_setopt_array($ch3, $options);
curl_exec($ch3);
curl_close($ch3);

$ch4 = curl_init("https://facebook.com:443");
curl_setopt($ch4, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch4, CURLOPT_TIMEOUT, 1);
curl_exec($ch4);
curl_close($ch4);

$ch5 = curl_init("http://www.aikido.dev:80");
curl_setopt($ch5, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch5, CURLOPT_TIMEOUT, 1);
curl_exec($ch5);
curl_close($ch5);

$ch6 = curl_init("http://some-invalid-domain.com:4113");
curl_setopt($ch6, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch6, CURLOPT_TIMEOUT, 1);
curl_exec($ch6);
curl_close($ch6);

?>

--EXPECT--
[AIKIDO][INFO] [BEFORE] Got domain: example.com
[AIKIDO][INFO] [AFTER] Got domain: example.com port: 443
[AIKIDO][INFO] [BEFORE] Got domain: httpbin.org
[AIKIDO][INFO] [AFTER] Got domain: httpbin.org port: 443
[AIKIDO][INFO] [BEFORE] Got domain: facebook.com
[AIKIDO][INFO] [AFTER] Got domain: facebook.com port: 443
[AIKIDO][INFO] [BEFORE] Got domain: facebook.com
[AIKIDO][INFO] [AFTER] Got domain: facebook.com port: 443
[AIKIDO][INFO] [BEFORE] Got domain: www.aikido.dev
[AIKIDO][INFO] [AFTER] Got domain: www.aikido.dev port: 80
[AIKIDO][INFO] [BEFORE] Got domain: some-invalid-domain.com
[AIKIDO][INFO] [AFTER] Got domain: some-invalid-domain.com port: 4113