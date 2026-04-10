<?php

$pdo = new PDO("sqlite::memory:");
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE IF NOT EXISTS addresses (
            id INTEGER PRIMARY KEY,
            name TEXT,
            street TEXT,
            city TEXT,
            zip TEXT,
            country TEXT)");

$requestBody = file_get_contents('php://input');
$data = json_decode($requestBody, true);

$stmt = $pdo->prepare("INSERT INTO addresses (name, street, city, zip, country) VALUES (:name, :street, :city, :zip, :country)");
$stmt->execute([
    'name' => $data['title'],
    'street' => '123 Main St',
    'city' => 'Springfield',
    'zip' => '12345',
    'country' => 'US',
]);

echo "Query executed!";

?>
