--TEST--
Test IDOR protection resolves positional "?" placeholders passed to PDOStatement::execute()

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT)");

$insert = $pdo->prepare("INSERT INTO cats (name, tenant_id) VALUES (?, ?)");
$insert->execute(['Mittens', 'org_123']);
echo "insert ok\n";

$select = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = ?");
$select->execute(['org_123']);
echo "select ok: " . $select->fetchColumn() . "\n";

$update = $pdo->prepare("UPDATE cats SET name = ? WHERE tenant_id = ?");
$update->execute(['Tom', 'org_123']);
echo "update ok\n";

?>

--EXPECTREGEX--
.*insert ok.*select ok: Mittens.*update ok.*
