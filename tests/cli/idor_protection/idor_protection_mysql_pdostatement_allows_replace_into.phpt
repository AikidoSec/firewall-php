--TEST--
Test IDOR protection allows a MySQL REPLACE INTO with the correct tenant column and value

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

$replace = $pdo->prepare("REPLACE INTO cats (name, tenant_id) VALUES (?, ?)");
$replace->execute(['Mittens', 'org_123']);
echo "replace ok\n";

$replace->execute(['Mittens', 'org_123']);
echo "replace again ok\n";

$select = $pdo->prepare("SELECT COUNT(*) FROM cats WHERE tenant_id = ?");
$select->execute(['org_123']);
echo "rows: " . $select->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*replace ok.*replace again ok.*rows: 1.*
