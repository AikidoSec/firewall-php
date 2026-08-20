--TEST--
Test IDOR protection blocks an INSERT with a positional "?" placeholder that sets the wrong tenant ID value

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

// The tenant column is the second placeholder: its value must be read from the
// second bound parameter, not the first one.
$insert = $pdo->prepare("INSERT INTO cats (name, tenant_id) VALUES (?, ?)");
$insert->execute(['org_123', 'org_456']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: INSERT on table 'cats' sets 'tenant_id' to 'org_456' but tenant ID is 'org_123'.*
