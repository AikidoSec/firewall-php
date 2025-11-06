<?php
    
\aikido\set_user("12345", "TestUser");

// Simple endpoint for the server to call itself
if ($_SERVER['REQUEST_URI'] === '/test') {
    echo "Test endpoint response";
    exit;
}

// Read the raw POST body
$requestBody = file_get_contents('php://input');

// Decode the JSON data to an associative array
$data = json_decode($requestBody, true);

if (isset($data['url'])) {
    $ch1 = curl_init($data['url']);
    curl_setopt($ch1, CURLOPT_RETURNTRANSFER, true);
    // Allow self-signed certificates for testing
    curl_setopt($ch1, CURLOPT_SSL_VERIFYPEER, false);
    curl_setopt($ch1, CURLOPT_SSL_VERIFYHOST, false);
    // Very short timeout to avoid hanging when requesting itself
    // (PHP built-in server is single-threaded and would deadlock)
    curl_setopt($ch1, CURLOPT_TIMEOUT_MS, 3000); // 3 second timeout
    //curl_setopt($ch1, CURLOPT_CONNECTTIMEOUT_MS, 3000); // 3 second connect timeout
    try {
        $result = curl_exec($ch1);
    } catch (Exception $e) {
        http_response_code(500);
        echo "Error: " . $e->getMessage();
        exit;
    }
    curl_close($ch1);
    
    // The important part: if we reach here, Aikido didn't block it as SSRF
    // Even if curl timed out trying to connect to itself, that's expected
    echo "Got URL content!";
 
} else {
    echo "Field 'url' is not present in the JSON data.";
}

?>

