<?php

\aikido\set_user("12345", "Tudor");

$mysqli = new mysqli("127.0.0.1", "root", "pwd", "db");

if ($mysqli->connect_error) {
    die("Connection failed: " . $mysqli->connect_error);
}

$createTable = "
    CREATE TEMPORARY TABLE IF NOT EXISTS users (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255),
        email VARCHAR(255)
    )
";
$mysqli->query($createTable);

$mysqli->query("INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')");

$requestBody = file_get_contents('php://input');
$data = json_decode($requestBody, true);

$userId = $data['userId'];
$result = $mysqli->query("SELECT * FROM users WHERE id = " . $userId);

echo "Query executed!";

$mysqli->close();

?>
