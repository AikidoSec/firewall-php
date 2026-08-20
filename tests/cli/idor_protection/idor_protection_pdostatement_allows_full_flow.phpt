--TEST--
Test IDOR protection allows a full flow with correct tenant filters (PDOStatement::execute)

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id", ["migrations"]);
\aikido\set_tenant_id("org_123");

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT)");
$pdo->exec("CREATE TABLE migrations (id INTEGER)");

echo "tenant is: " . \aikido\get_tenant_id() . "\n";

$insert = $pdo->prepare("INSERT INTO cats (name, tenant_id) VALUES (:name, :tenant_id)");
$insert->execute([':name' => 'Mittens', ':tenant_id' => 'org_123']);
echo "insert ok\n";

$select = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = :tenant_id");
$select->execute([':tenant_id' => 'org_123']);
echo "select ok: " . $select->fetchColumn() . "\n";

$excluded = $pdo->prepare("SELECT COUNT(*) FROM migrations");
$excluded->execute();
echo "excluded table ok\n";

?>

--EXPECTREGEX--
.*tenant is: org_123.*insert ok.*select ok: Mittens.*excluded table ok.*
