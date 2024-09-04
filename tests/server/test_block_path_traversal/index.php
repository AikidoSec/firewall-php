<?php
    
// Read the raw POST body
$requestBody = file_get_contents('php://input');

// Decode the JSON data to an associative array
$data = json_decode($requestBody, true);

// Check if 'folder' exists and get its value
if (isset($data['folder'])) {
    $f = $data['folder'] . '/file';
    echo "Opening fileFolder: " . $f;
    fopen($f, 'r');
} else {
    echo "Field 'folder' is not present in the JSON data.";
}

?>
