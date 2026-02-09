--TEST--
Test SSRF with CURLOPT_HEADERFUNCTION callback (process response headers)

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

    // Perform the cURL request with HEADERFUNCTION callback
    $ch1 = curl_init("https://ssrf-redirects.testssandbox.com/ssrf-test-4");
    curl_setopt($ch1, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    
    // CURLOPT_HEADERFUNCTION: Processes response headers
    // This callback triggers nested hooks via file operations
    curl_setopt($ch1, CURLOPT_HEADERFUNCTION, function($ch, $header) {
        // Nested hook invocation
        $fp = @fopen('/tmp/curl_headers.log', 'a');
        if ($fp) {
            fwrite($fp, $header);
            fclose($fp);
        }
        return strlen($header);
    });
    
    $response = curl_exec($ch1);
    curl_close($ch1);

    echo "Response received:\n";
    if (file_exists('/tmp/curl_headers.log')) {
        echo "Headers logged\n";
        unlink('/tmp/curl_headers.log');
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

