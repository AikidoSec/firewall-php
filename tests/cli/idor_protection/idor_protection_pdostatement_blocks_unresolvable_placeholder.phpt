--TEST--
Test IDOR protection blocks a query whose tenant placeholder cannot be resolved

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

// A NULL tenant ID can never match the configured tenant, and we do not want to
// silently compare it against the empty string either.
$stmt = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id");
$stmt->execute([':tenant_id' => null]);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' has a placeholder for 'tenant_id' that could not be resolved.*
