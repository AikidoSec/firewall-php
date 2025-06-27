--TEST--
Test PDO::exec() method for SQL injection
--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--GET--
name=ls%F0%28%8C%BC&age='||sqlite_version()||'

--FILE--
<?php

try {

    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
   
    $pdo->exec("CREATE TABLE IF NOT EXISTS cats (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age TEXT NOT NULL
    )");


    $pdo->exec("INSERT INTO cats (name, age) VALUES ( 'name', '" . $_GET['age'] . "')");

    echo "Query executed!";
    var_dump($pdo->query("SELECT * FROM cats")->fetchObject());

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
