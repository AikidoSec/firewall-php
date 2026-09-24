--TEST--
Test IDOR protection allows PDO::query() SELECT with a correct tenant filter

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

$result = $pdo->query("SELECT name FROM cats WHERE tenant_id = 'org_123'");
echo "select ok: " . $result->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*select ok: Mittens.*
