--TEST--
Test SSRF with multiple callbacks combined (comprehensive nested hook stress test)

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

    // Perform the cURL request with MULTIPLE callbacks simultaneously
    // This tests that the EventCacheStack properly handles multiple nested hooks
    // from different callback types in the same request
    $ch1 = curl_init("https://ssrf-redirects.testssandbox.com/ssrf-test-4");
    curl_setopt($ch1, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    
    // CURLOPT_WRITEFUNCTION: Process response body
    curl_setopt($ch1, CURLOPT_WRITEFUNCTION, function($ch, $data) {
        // Nested hook 1 - file_put_contents
        file_put_contents('/tmp/curl_multi_write.log', strlen($data) . " bytes\n", FILE_APPEND);
        return strlen($data);
    });
    
    // CURLOPT_HEADERFUNCTION: Process response headers
    curl_setopt($ch1, CURLOPT_HEADERFUNCTION, function($ch, $header) {
        // Nested hook 2 - fopen/fwrite/fclose chain
        $fp = @fopen('/tmp/curl_multi_headers.log', 'a');
        if ($fp) {
            fwrite($fp, strlen($header) . " header bytes\n");
            fclose($fp);
        }
        return strlen($header);
    });
    
    $response = curl_exec($ch1);
    curl_close($ch1);

    echo "Response received:\n";
    if (file_exists('/tmp/curl_multi_write.log')) {
        unlink('/tmp/curl_multi_write.log');
    }
    if (file_exists('/tmp/curl_multi_headers.log')) {
        unlink('/tmp/curl_multi_headers.log');
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

