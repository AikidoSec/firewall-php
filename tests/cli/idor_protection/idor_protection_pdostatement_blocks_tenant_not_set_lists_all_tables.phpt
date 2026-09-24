--TEST--
Test IDOR protection lists every non-excluded table when \aikido\set_tenant_id() was never called

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE tickets (id INTEGER PRIMARY KEY AUTOINCREMENT, tenant_id TEXT)");
$pdo->exec("CREATE TABLE comments (id INTEGER PRIMARY KEY AUTOINCREMENT, ticket_id INTEGER, tenant_id TEXT)");

\aikido\enable_idor_protection("tenant_id");

$stmt = $pdo->prepare("SELECT t.id FROM tickets t JOIN comments c ON c.ticket_id = t.id");
$stmt->execute();

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on tables 'tickets, comments' requires a tenant ID, but set_tenant_id\(\) was not called\..*
