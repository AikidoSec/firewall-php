<?php

# if path has /api/execute/<command>, execute the command

$path = $_SERVER['REQUEST_URI'];
if (strpos($path, '/api/execute/') === 0) {

    $command = substr($path, strlen('/api/execute/'));
    passthru($command);
    echo "Shell executed!";

} else {


    \aikido\set_user("12345", "Tudor");

    // Read the raw POST body
    $requestBody = file_get_contents('php://input');

    // Decode the JSON data to an associative array
    $data = json_decode($requestBody, true);

    if (isset($data['command'])) {
        passthru('binary --domain www.example' .  $data['command'] . '.com');
        echo "Shell executed!";
    } else {
        echo "Field 'command' is not present in the JSON data.";
    }
}
?>
