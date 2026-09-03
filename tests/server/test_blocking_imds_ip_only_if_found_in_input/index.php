<?php
    
$get_instance_metadata_id = function($url) {
    $url = 'http://' . $url . '/tests/latest/meta-data/instance-id';
    
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_HTTPHEADER, ['X-aws-ec2-metadata-token: ' . 'test-token']);
    curl_setopt($ch, CURLOPT_TIMEOUT_MS, 5);

    curl_exec($ch);
    
    return "test_instance_id";
};

// Read the raw POST body
$requestBody = file_get_contents('php://input');

// Decode the JSON data to an associative array
$data = json_decode($requestBody, true);

$url = "169.254.169.254";
if (isset($data['url'])) {
    $url = $data['url'];
}

echo "Instance id: " . $get_instance_metadata_id($url) . "\n";

?>
