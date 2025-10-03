--TEST--
Request https://ssrf-redirects.testssandbox.com/ssrf-test-4 with CURLOPT_FOLLOWLOCATION = false to inspect the Location: header and then manually follow redirects — mirrors how Guzzle, Symfony HttpClient auto-follow redirects.
--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

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
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch1, CURLOPT_FOLLOWLOCATION, false);

    $response = curl_exec($ch1);

    if ($response === false) {
        echo "cURL Error: " . curl_error($ch);
    } else {
        echo "Response: " . $response . "\n";
        $code = curl_getinfo($ch1, CURLINFO_HTTP_CODE);
        echo "Code: " . $code . "\n";
        $redirectUrl = curl_getinfo($ch1, CURLINFO_REDIRECT_URL);
        echo "Redirect URL: " . $redirectUrl . "\n";

        $ch2 = curl_init($redirectUrl);
        curl_setopt($ch2, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch2, CURLOPT_FOLLOWLOCATION, true);
        $response2 = curl_exec($ch2);
        echo "Response2: " . $response2;
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
