--TEST--
Test IDOR protection blocks a MySQL INSERT ... AS alias ON DUPLICATE KEY UPDATE upsert that sets the wrong tenant ID value

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
$pdo->exec("CREATE TEMPORARY TABLE cats (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), tenant_id VARCHAR(255), UNIQUE KEY uniq_name (name))");

$upsert = $pdo->prepare("INSERT INTO cats (name, tenant_id) VALUES (?, ?) AS vals ON DUPLICATE KEY UPDATE name = vals.name, tenant_id = vals.tenant_id");
$upsert->execute(['Mittens', 'org_456']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: INSERT on table 'cats' sets 'tenant_id' to 'org_456' but tenant ID is 'org_123'.*
