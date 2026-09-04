<?php

require_once __DIR__ . '/get_instance_metadata_id.php';

// Read the raw POST body
$requestBody = file_get_contents('php://input');

// Decode the JSON data to an associative array
$data = json_decode($requestBody, true);

$url = "169.254.169.254";
if (isset($data['url'])) {
    $url = $data['url'];
}

echo "Instance id: " . get_instance_metadata_id($url) . "\n";

?>
