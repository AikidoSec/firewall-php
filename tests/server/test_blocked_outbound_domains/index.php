<?php

// Read the raw POST body
$requestBody = file_get_contents('php://input');

// Decode the JSON data to an associative array
$data = json_decode($requestBody, true);

if (isset($data['url'])) {
    try {
        // Make an outbound request using curl
        $ch = curl_init($data['url']);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, 5);
        $result = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);
        
        if ($result === false) {
            echo "Request failed";
        } else {
            echo "Request succeeded: " . $httpCode;
        }
    } catch (Exception $e) {
        // Catch Aikido blocking exception
        http_response_code(500);
        echo "Request blocked by Aikido: " . $e->getMessage();
    }
} else {
    echo "Field 'url' is not present in the JSON data.";
}

?>

