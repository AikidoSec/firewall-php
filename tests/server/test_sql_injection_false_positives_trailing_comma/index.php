<?php

$pdo = new PDO("sqlite::memory:");
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY,
            name TEXT,
            email TEXT,
            status TEXT)");

$pdo->exec("INSERT INTO users (name, email, status) VALUES ('John Doe', 'john@example.com', 'active')");

$requestBody = file_get_contents('php://input');
$data = json_decode($requestBody, true);

$stmt = $pdo->prepare("SELECT * FROM users WHERE name = :name AND email IS NOT NULL AND status NOT IN ('SUSPENDED', 'DELETED')");
$stmt->execute(['name' => $data['name']]);

echo "Query executed!";

?>
