<?php

\aikido\set_user("12345", "Tudor");

// Connect to MySQL (adjust credentials as needed)
$conn = mysqli_connect("localhost", "root", "", "db");

// Check connection
if (!$conn) {
    die("Connection failed: " . mysqli_connect_error());
}

// Create a temporary table
$createTable = "
    CREATE TEMPORARY TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255),
        email VARCHAR(255)
    )
";
mysqli_query($conn, $createTable);

// Insert a row
mysqli_query($conn, "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')");

// Read raw POST input
$requestBody = file_get_contents('php://input');

// Decode JSON
$data = json_decode($requestBody, true);

// Vulnerable query
$userId = $data['userId'];
$result = mysqli_real_query($conn, "SELECT * FROM users WHERE id = " . $userId);

echo "Query executed!";

// Close the connection
mysqli_close($conn);

?>
