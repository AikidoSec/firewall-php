<?php

\aikido\set_user("12345", "Tudor");

// Connect to MySQL (adjust credentials as needed)
$mysqli = new mysqli("127.0.0.1", "root", "pwd", "db");

// Check connection
if ($mysqli->connect_error) {
    die("Connection failed: " . $mysqli->connect_error);
}

// Create a temporary table
$createTable = "
    CREATE TEMPORARY TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255),
        email VARCHAR(255)
    )
";
$mysqli->query($createTable);

// Insert a row
$mysqli->query("INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')");

// Read raw POST input
$requestBody = file_get_contents('php://input');

// Decode JSON
$data = json_decode($requestBody, true);

// Vulnerable query
$userId = $data['userId'];
$result = $mysqli->real_query("SELECT * FROM users WHERE id = " . $userId);

echo "Query executed!";

// Close the connection
$mysqli->close();

?>
