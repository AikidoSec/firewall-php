--TEST--
Test IDOR protection uses the PostgreSQL dialect for PDO PostgreSQL queries

--SKIPIF--
<?php
if (!extension_loaded('pdo_pgsql')) {
    die('skip pdo_pgsql not available');
}
try {
    new PDO('pgsql:host=127.0.0.1;port=5432;dbname=zen_test', 'postgres', 'postgres');
} catch (PDOException $e) {
    die('skip PostgreSQL server not reachable on 127.0.0.1');
}
?>

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection('tenant_id');
\aikido\set_tenant_id('org_123');

$pdo = new PDO('pgsql:host=127.0.0.1;port=5432;dbname=zen_test', 'postgres', 'postgres');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec('CREATE TEMPORARY TABLE cats (id SERIAL PRIMARY KEY, name TEXT, tenant_id TEXT)');
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

$stmt = $pdo->prepare('SELECT name FROM cats WHERE tenant_id = ?::text');
$stmt->execute(['org_456']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' filters 'tenant_id' with value 'org_456' but tenant ID is 'org_123'.*
