--TEST--
Test IDOR protection is skipped entirely when \aikido\enable_idor_protection() was never called

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT)");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

$stmt = $pdo->prepare("SELECT name FROM cats");
$stmt->execute();
echo "select ok: " . $stmt->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*select ok: Mittens.*
