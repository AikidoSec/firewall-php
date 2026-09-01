--TEST--
Test \aikido\without_idor_protection() bypasses the IDOR check for the wrapped query

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, tenant_id TEXT)");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Rex', 'org_456')");

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$count = \aikido\without_idor_protection(function () use ($pdo) {
    $stmt = $pdo->prepare("SELECT COUNT(*) FROM cats");
    $stmt->execute();
    return $stmt->fetchColumn();
});

echo "count across tenants: " . $count . "\n";

?>

--EXPECTREGEX--
.*count across tenants: 2.*
