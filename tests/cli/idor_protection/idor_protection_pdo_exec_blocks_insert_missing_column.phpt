--TEST--
Test IDOR protection blocks PDO::exec() INSERT missing the tenant column

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

$pdo->exec("INSERT INTO cats (name) VALUES ('Mittens')");

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: INSERT on table 'cats' is missing column 'tenant_id'.*
