--TEST--
Test socket functions (socket_connect, fsockopen, stream_socket_client)

--SKIPIF--
<?php
if (PHP_VERSION_ID >= 80500) {
    die("skip PHP >= 8.5.");
}
if (!extension_loaded('sockets')) {
    die("skip sockets extension not loaded");
}
?>

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php
// Test socket_connect
$socket1 = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
if ($socket1 !== false) {
    @socket_connect($socket1, "example.com", 443);
    socket_close($socket1);
}

// Test fsockopen
$fp1 = @fsockopen("httpbin.org", 443, $errno, $errstr, 1);
if ($fp1) {
    fclose($fp1);
}

// Test stream_socket_client
$fp2 = @stream_socket_client("tcp://facebook.com:443", $errno, $errstr, 1);
if ($fp2) {
    fclose($fp2);
}

// Test socket_connect with explicit port
$socket2 = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
if ($socket2 !== false) {
    @socket_connect($socket2, "www.aikido.dev", 80);
    socket_close($socket2);
}

// Test fsockopen with explicit port
$fp3 = @fsockopen("some-invalid-domain.com", 4113, $errno, $errstr, 1);
if ($fp3) {
    fclose($fp3);
}

// Test stream_socket_client with http:// scheme
$fp4 = @stream_socket_client("http://www.aikido.dev:80", $errno, $errstr, 1);
if ($fp4) {
    fclose($fp4);
}

// Test stream_socket_client with https:// scheme
$fp5 = @stream_socket_client("https://example.com:443", $errno, $errstr, 1);
if ($fp5) {
    fclose($fp5);
}

?>

--EXPECT--
[AIKIDO][INFO] [BEFORE] Got domain: example.com
[AIKIDO][INFO] [AFTER] Got domain: example.com port: 443
[AIKIDO][INFO] [BEFORE] Got domain: httpbin.org
[AIKIDO][INFO] [AFTER] Got domain: httpbin.org port: 443
[AIKIDO][INFO] [BEFORE] Got domain: facebook.com
[AIKIDO][INFO] [AFTER] Got domain: facebook.com port: 443
[AIKIDO][INFO] [BEFORE] Got domain: www.aikido.dev
[AIKIDO][INFO] [AFTER] Got domain: www.aikido.dev port: 80
[AIKIDO][INFO] [BEFORE] Got domain: some-invalid-domain.com
[AIKIDO][INFO] [AFTER] Got domain: some-invalid-domain.com port: 4113
[AIKIDO][INFO] [BEFORE] Got domain: www.aikido.dev
[AIKIDO][INFO] [AFTER] Got domain: www.aikido.dev port: 80
[AIKIDO][INFO] [BEFORE] Got domain: example.com
[AIKIDO][INFO] [AFTER] Got domain: example.com port: 443

