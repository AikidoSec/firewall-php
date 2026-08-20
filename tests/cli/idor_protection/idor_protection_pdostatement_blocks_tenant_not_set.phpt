--TEST--
Test IDOR protection blocks a query when \aikido\set_tenant_id() was never called

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT)");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

\aikido\enable_idor_protection("tenant_id");

$stmt = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id");
$stmt->execute([':tenant_id' => 'org_123']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' requires a tenant ID, but set_tenant_id\(\) was not called\..*
