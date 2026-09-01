--TEST--
Test IDOR protection allows a MySQL INSERT ... VALUES ... AS alias ON DUPLICATE KEY UPDATE upsert with the correct tenant value

--SKIPIF--
<?php
if (!extension_loaded("pdo_mysql")) {
    die("skip pdo_mysql not available");
}
try {
    $probe = new PDO("mysql:host=127.0.0.1;dbname=db", "root", "pwd");
    $probe->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
} catch (PDOException $e) {
    die("skip MySQL server not reachable on 127.0.0.1");
}
try {
    $probe->exec("CREATE TEMPORARY TABLE zen_row_alias_probe (id INT PRIMARY KEY)");
    $probe->exec("INSERT INTO zen_row_alias_probe (id) VALUES (1) AS vals ON DUPLICATE KEY UPDATE id = vals.id");
} catch (PDOException $e) {
    die("skip server does not support the VALUES row alias (needs MySQL 8.0.19+)");
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

// The tenant column is checked on the VALUES clause, not on the
// ON DUPLICATE KEY UPDATE assignments.
$upsert = $pdo->prepare("INSERT INTO cats (name, tenant_id) VALUES (?, ?) AS vals ON DUPLICATE KEY UPDATE name = vals.name, tenant_id = vals.tenant_id");
$upsert->execute(['Mittens', 'org_123']);
echo "insert ok\n";

// Same statement again, this time hitting the duplicate key and updating.
$upsert->execute(['Mittens', 'org_123']);
echo "upsert ok\n";

$select = $pdo->prepare("SELECT COUNT(*) FROM cats WHERE tenant_id = ?");
$select->execute(['org_123']);
echo "rows: " . $select->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*insert ok.*upsert ok.*rows: 1.*
