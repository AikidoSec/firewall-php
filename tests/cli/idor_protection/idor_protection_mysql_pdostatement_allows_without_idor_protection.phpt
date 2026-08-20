--TEST--
Test \aikido\without_idor_protection() bypasses the IDOR check for a MySQL query

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

$pdo = new PDO('mysql:host=127.0.0.1;dbname=db', 'root', 'pwd');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TEMPORARY TABLE cats (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), tenant_id VARCHAR(255))");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Rex', 'org_456')");

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$count = \aikido\without_idor_protection(function () use ($pdo) {
    $stmt = $pdo->prepare("SELECT COUNT(*) FROM cats");
    $stmt->execute();
    return $stmt->fetchColumn();
});

echo "count across tenants: " . $count . "\n";

?>

--EXPECTREGEX--
.*count across tenants: 2.*
