<?php

\aikido\set_user("12345", "Test User");

try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    $pdo->exec("CREATE TABLE IF NOT EXISTS cats (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age INTEGER NOT NULL
    )");

    $pdo->exec("INSERT INTO cats (name, age) VALUES ('n', 1)");

    $id = $_GET['id'];
    $pdo->query("SELECT * FROM cats WHERE id = $id");

    echo "Query executed!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
