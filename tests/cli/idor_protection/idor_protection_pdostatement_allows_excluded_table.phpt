--TEST--
Test IDOR protection allows queries on excluded tables without a tenant filter

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id", ["migrations"]);
\aikido\set_tenant_id("org_123");

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE migrations (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)");
$pdo->exec("INSERT INTO migrations (name) VALUES ('add_cats_table')");

$stmt = $pdo->prepare("SELECT name FROM migrations");
$stmt->execute();
echo "excluded table query ok: " . $stmt->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*excluded table query ok: add_cats_table.*
