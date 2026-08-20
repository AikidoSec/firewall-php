--TEST--
Test IDOR protection allows migrations, health checks and excluded-table lookups before \aikido\set_tenant_id() is called

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

// The documented startup order: turn protection on first, then bootstrap the
// app. None of the queries below need a tenant, so they must run before one is
// set -- the lookup in "users" is what tells us which tenant to set.
\aikido\enable_idor_protection("tenant_id", ["users"]);

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

$pdo->exec("CREATE TABLE IF NOT EXISTS users (id INTEGER, email TEXT)");
$pdo->exec("CREATE TABLE IF NOT EXISTS cats (id INTEGER, name TEXT, tenant_id TEXT)");
$pdo->exec("CREATE INDEX IF NOT EXISTS cats_tenant ON cats (tenant_id)");
echo "migrations ok\n";

$pdo->exec("BEGIN");
$pdo->exec("COMMIT");
echo "transaction ok\n";

echo "health check: " . $pdo->query("SELECT 1")->fetchColumn() . "\n";

$pdo->exec("INSERT INTO users (id, email) VALUES (1, 'alice@example.com')");
$lookup = $pdo->prepare("SELECT id FROM users WHERE email = ?");
$lookup->execute(['alice@example.com']);
echo "excluded table lookup: " . $lookup->fetchColumn() . "\n";

// Only now do we know the tenant.
\aikido\set_tenant_id("org_123");
$insert = $pdo->prepare("INSERT INTO cats (name, tenant_id) VALUES (?, ?)");
$insert->execute(['Mittens', 'org_123']);
echo "insert ok\n";

// A tenant-scoped table still needs the tenant, and it is set now.
$select = $pdo->prepare("SELECT name FROM cats WHERE tenant_id = ?");
$select->execute(['org_123']);
echo "select ok: " . $select->fetchColumn() . "\n";

?>

--EXPECTREGEX--
.*migrations ok.*transaction ok.*health check: 1.*excluded table lookup: 1.*insert ok.*select ok: Mittens.*
