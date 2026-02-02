--TEST--
Test SSRF with CURLOPT_PROGRESSFUNCTION callback (monitor transfer progress)

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

    // Perform the cURL request with PROGRESSFUNCTION callback
    $ch1 = curl_init("https://ssrf-redirects.testssandbox.com/ssrf-test-4");
    curl_setopt($ch1, CURLOPT_FOLLOWLOCATION, true);
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch1, CURLOPT_NOPROGRESS, false); // Enable progress function
    
    // CURLOPT_PROGRESSFUNCTION: Monitors transfer progress
    // This callback triggers nested hooks via file operations
    curl_setopt($ch1, CURLOPT_PROGRESSFUNCTION, function($resource, $download_size, $downloaded, $upload_size, $uploaded) {
        // Nested hook invocation 
        if ($downloaded > 0 || $uploaded > 0) {
            $progress = sprintf("DL: %d/%d, UL: %d/%d\n", $downloaded, $download_size, $uploaded, $upload_size);
            file_put_contents('/tmp/curl_progress.log', $progress, FILE_APPEND);
        }
        return 0; // Return 0 to continue, non-zero to abort
    });
    
    $response = curl_exec($ch1);
    curl_close($ch1);

    echo "Response received:\n";
    if (file_exists('/tmp/curl_progress.log')) {
        echo "Progress logged\n";
        unlink('/tmp/curl_progress.log');
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

