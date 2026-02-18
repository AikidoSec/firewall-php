<?php
    
\aikido\set_user("12345", "Tudor");

// Read the raw POST body
$requestBody = file_get_contents('php://input');

// Decode the JSON data to an associative array
$data = json_decode($requestBody, true);

if (isset($data['url'])) {
    $parsedUrl = parse_url($data['url']);
    $host = $parsedUrl['host'] ?? '127.0.0.1';
    $port = $parsedUrl['port'] ?? 80;
    
    // Test socket_connect
    if (isset($data['method']) && $data['method'] === 'socket_connect') {
        $socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
        if ($socket !== false) {
            socket_connect($socket, $host, $port);
            socket_close($socket);
        }
        echo "Socket connect completed!";
    }
    // Test fsockopen
    elseif (isset($data['method']) && $data['method'] === 'fsockopen') {
        $fp = @fsockopen($host, $port, $errno, $errstr, 5);
        if ($fp) {
            fclose($fp);
            echo "Fsockopen completed!";
        } else {
            echo "Fsockopen failed: $errstr ($errno)";
        }
    }
    // Test stream_socket_client
    elseif (isset($data['method']) && $data['method'] === 'stream_socket_client') {
        $context = stream_context_create(['socket' => ['connect_timeout' => 5]]);
        $fp = @stream_socket_client($data['url'], $errno, $errstr, 5, STREAM_CLIENT_CONNECT, $context);
        if ($fp) {
            fclose($fp);
            echo "Stream socket client completed!";
        } else {
            echo "Stream socket client failed: $errstr ($errno)";
        }
    }
    // Default: test all socket methods
    else {
        // Test socket_connect
        $socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
        if ($socket !== false) {
            @socket_connect($socket, $host, $port);
            socket_close($socket);
        }
        
        // Test fsockopen
        $fp = @fsockopen($host, $port, $errno, $errstr, 5);
        if ($fp) {
            fclose($fp);
        }
        
        // Test stream_socket_client
        $context = stream_context_create(['socket' => ['connect_timeout' => 5]]);
        $fp = @stream_socket_client($data['url'], $errno, $errstr, 5, STREAM_CLIENT_CONNECT, $context);
        if ($fp) {
            fclose($fp);
        }
        
        echo "All socket methods completed!";
    }
} else {
    echo "Field 'url' is not present in the JSON data.";
}

?>

