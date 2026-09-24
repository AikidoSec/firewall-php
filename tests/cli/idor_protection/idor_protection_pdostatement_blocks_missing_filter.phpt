--TEST--
Test IDOR protection blocks PDOStatement::execute() SELECT without a tenant filter, even in detection-only mode

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=0

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT)");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

$stmt = $pdo->prepare("SELECT name FROM cats");
$stmt->execute();

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' is missing a filter on column 'tenant_id'.*
