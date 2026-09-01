--TEST--
Test IDOR protection allows MySQL queries on excluded tables without a tenant filter

--SKIPIF--
<?php
if (!extension_loaded("pdo_mysql")) {
    die("skip pdo_mysql not available");
}
try {
    new PDO("mysql:host=127.0.0.1;dbname=db", "root", "pwd");
} catch (PDOException $e) {
    die("skip MySQL server not reachable on 127.0.0.1");
}
?>

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id", ["migrations"]);
\aikido\set_tenant_id("org_123");

$pdo = new PDO('mysql:host=127.0.0.1;dbname=db', 'root', 'pwd');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TEMPORARY TABLE migrations (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255))");
$pdo->exec("INSERT INTO migrations (name) VALUES ('add_cats_table')");

$stmt = $pdo->prepare("SELECT name FROM migrations");
$stmt->execute();
echo "excluded table query ok: " . $stmt->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*excluded table query ok: add_cats_table.*
