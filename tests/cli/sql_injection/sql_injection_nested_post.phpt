--TEST--
Test SQL injection detection in nested POST parameters (filters[id][value])

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--POST--
filters[id][value]=1%20OR%201%3D1--

--FILE--
<?php

try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    $pdo->exec("CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY,
        name TEXT NOT NULL
    )");

    $pdo->exec("INSERT INTO users VALUES (1, 'admin')");
    $pdo->exec("INSERT INTO users VALUES (2, 'user')");

    $id = $_POST['filters']['id']['value'];
    $result = $pdo->query("SELECT * FROM users WHERE id=" . $id);

    echo "Query executed!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
