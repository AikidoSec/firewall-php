--TEST--
Ensure cURL requests to example.com and after to a local dev server (127.0.0.1) are not incorrectly blocked as SSRF by Aikido

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--POST--
test=https://example.com

--FILE--
<?php

$host = '127.0.0.1';
$port = 3000;
$pid = null;

$descriptorspec = [
    0 => ['pipe', 'r'],
    1 => ['file', '/dev/null', 'a'],
    2 => ['file', '/dev/null', 'a']
];

try {
    // Start PHP server
    $process = proc_open("php -S $host:$port", $descriptorspec, $pipes);
    if (!is_resource($process)) {
        throw new RuntimeException("Failed to start PHP server.");
    }

    $status = proc_get_status($process);
    $pid = $status['pid'];

    // Wait a moment to ensure server starts
    sleep(1);

    // Perform the cURL request
    $ch1 = curl_init("https://example.com");
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch1, CURLOPT_FOLLOWLOCATION, true);

    $response = curl_exec($ch1);

    if ($response === false) {
        echo "cURL Error: " . curl_error($ch);
    } else {
        echo "Response: " . $response;
    }

    $ch2 = curl_init("http://127.0.0.1:3000");
    curl_setopt($ch2, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch2, CURLOPT_FOLLOWLOCATION, true);
    $response2 = curl_exec($ch2);


} catch (Throwable $e) {
    echo "Error: " . $e->getMessage() . "\n";
} finally {
    // Ensure the server is killed if started
    if ($pid) {
        exec("kill -9 $pid");
    }
    if (isset($process) && is_resource($process)) {
        proc_close($process);
    }
}


--EXPECTREGEX--
(?s)\A(?!.*Aikido firewall has blocked a server-side request forgery).*?\z
