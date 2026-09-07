--TEST--
Test IDOR protection still runs when another bound parameter holds non-UTF-8 bytes

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec("CREATE TABLE cats (id INTEGER PRIMARY KEY AUTOINCREMENT, picture BLOB, tenant_id TEXT)");

// Serializing the bound values must never throw on invalid UTF-8: that would
// abort the hook and silently skip the check for this query.
$stmt = $pdo->prepare("INSERT INTO cats (picture, tenant_id) VALUES (:picture, :tenant_id)");
$stmt->execute([':picture' => "\xff\xfe\x80binary", ':tenant_id' => 'org_456']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: INSERT on table 'cats' sets 'tenant_id' to 'org_456' but tenant ID is 'org_123'.*
