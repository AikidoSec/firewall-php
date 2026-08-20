--TEST--
Test IDOR protection allows a tenant filter combined with AND around an OR clause

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
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

$stmt = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id AND (name = :n1 OR name = :n2)");
$stmt->execute([':tenant_id' => 'org_123', ':n1' => 'Mittens', ':n2' => 'Felix']);
echo "select ok: " . $stmt->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*select ok: Mittens.*
