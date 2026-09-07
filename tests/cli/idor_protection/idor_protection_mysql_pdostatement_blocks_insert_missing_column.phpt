--TEST--
Test IDOR protection blocks PDOStatement::execute() INSERT on MySQL that omits the tenant column

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

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$pdo = new PDO('mysql:host=127.0.0.1;dbname=db', 'root', 'pwd');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TEMPORARY TABLE cats (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), tenant_id VARCHAR(255))");

$stmt = $pdo->prepare("INSERT INTO cats (name) VALUES (:name)");
$stmt->execute([':name' => 'Mittens']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: INSERT on table 'cats' is missing column 'tenant_id'.*
