--TEST--
Test SSRF with CURLOPT_WRITEFUNCTION callback (process response body chunks)

--ENV--
AIKIDO_BLOCK=1
AIKIDO_DEBUG=1

--POST--
test=https://ssrf-redirects.testssandbox.com/ssrf-test-4

--FILE--
<?php

$host = '127.0.0.1';
$port = 4000;
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
    $ch1 = curl_init("https://ssrf-redirects.testssandbox.com/ssrf-test-4");
    curl_setopt($ch1, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($ch1, CURLOPT_WRITEFUNCTION, function($ch, $data) {
        file_put_contents('/tmp/curl.log', $data, FILE_APPEND);
        return strlen($data);
    });
    $response = curl_exec($ch1);
    curl_close($ch1);

    echo "Response:\n";
    if (file_exists('/tmp/curl.log')) {
        echo file_get_contents('/tmp/curl.log');
        unlink('/tmp/curl.log');
    }

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
.*Aikido firewall has blocked a server-side request forgery.*
