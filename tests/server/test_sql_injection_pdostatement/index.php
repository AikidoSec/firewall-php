<?php
    
\aikido\set_user("12345", "Tudor");

try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
   
    $pdo->exec("CREATE TABLE IF NOT EXISTS cats (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age INTEGER NOT NULL
    )");


    $stmt = $pdo->prepare("INSERT INTO cats (name, age) VALUES ( :name, '" . $_GET['age'] . "')");
    $stmt->execute([':name' => $_GET['name']]);

    echo "Query executed!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>