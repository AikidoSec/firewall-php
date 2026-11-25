<?php

$hosts = [
    '/callback' => 'probe.oast.live',
    '/ftp' => 'probe.oast.live',
    '/rebinding' => 'probe.rbndr.us',
    '/wildcard' => 'test.sslip.io',
    '/blocked' => 'blocked.oast.live',
    '/lookalike' => 'notoast.live',
];
$route = parse_url($_SERVER['REQUEST_URI'], PHP_URL_PATH);
if (!isset($hosts[$route])) {
    http_response_code(404);
    return;
}

$host = $hosts[$route];
$port = (int) $_GET['port'];
$scheme = $route === '/ftp' ? 'ftp' : 'http';
$curl = curl_init("$scheme://$host:$port/mock/config_fetch_count");
curl_setopt_array($curl, [
    CURLOPT_RESOLVE => ["$host:$port:127.0.0.1"],
    CURLOPT_PROXY => '',
    CURLOPT_ENCODING => '',
    CURLOPT_TIMEOUT => 3,
    CURLOPT_RETURNTRANSFER => true,
]);
try {
    echo curl_exec($curl);
} catch (Exception $error) {
    http_response_code(500);
    echo $error->getMessage();
}
