<?php

// Handle API spec test route
if (strpos($_SERVER['REQUEST_URI'], '/api/v1/orders') !== false) {
    echo "Order processed!";
    exit;
}

// Handle SQL injection test route
try {
    $pdo = new PDO("sqlite::memory:");
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $pdo->exec("CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY, 
                name TEXT, 
                email TEXT)");

    $pdo->exec("INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')");

    // Read the raw POST body
    $requestBody = file_get_contents('php://input');

    // Decode the JSON data to an associative array
    $data = json_decode($requestBody, true);

    // Vulnerable query - SQL injection
    $result = $pdo->query("SELECT * FROM users WHERE id = " . $data['userId']);

    echo "Query executed!";
} catch (PDOException $e) {
    echo "Connection failed: " . $e->getMessage();
}

// Close the database connection
$pdo = null;

?>
