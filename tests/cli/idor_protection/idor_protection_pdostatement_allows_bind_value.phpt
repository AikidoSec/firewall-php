--TEST--
Test IDOR protection resolves values bound with bindValue()/bindParam() before execute()

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

// Named, with and without the leading colon (PDO accepts both).
$named = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id");
$named->bindValue(':tenant_id', 'org_123');
$named->execute();
echo "bindValue named ok: " . $named->fetchColumn() . "\n";

$unprefixed = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id");
$unprefixed->bindValue('tenant_id', 'org_123');
$unprefixed->execute();
echo "bindValue unprefixed ok: " . $unprefixed->fetchColumn() . "\n";

// Positional (bindValue takes a 1-based position).
$positional = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = ?");
$positional->bindValue(1, 'org_123');
$positional->execute();
echo "bindValue positional ok: " . $positional->fetchColumn() . "\n";

// bindParam binds by reference: the value at execute() time is what counts.
$byRef = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id");
$tenant = 'org_456';
$byRef->bindParam(':tenant_id', $tenant);
$tenant = 'org_123';
$byRef->execute();
echo "bindParam ok: " . $byRef->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*bindValue named ok: Mittens.*bindValue unprefixed ok: Mittens.*bindValue positional ok: Mittens.*bindParam ok: Mittens.*
