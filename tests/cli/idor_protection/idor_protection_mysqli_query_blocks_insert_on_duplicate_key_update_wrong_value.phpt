--TEST--
Test IDOR protection blocks a mysqli::query() INSERT ... AS alias ON DUPLICATE KEY UPDATE upsert with the wrong tenant ID value

--SKIPIF--
<?php
if (!extension_loaded("mysqli")) {
    die("skip mysqli not available");
}
try {
    $probe = @new mysqli("127.0.0.1", "root", "pwd", "db");
} catch (Exception $e) {
    die("skip MySQL server not reachable on 127.0.0.1");
}
if ($probe->connect_errno) {
    die("skip MySQL server not reachable on 127.0.0.1");
}
?>

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection("tenant_id");
\aikido\set_tenant_id("org_123");

$mysqli = new mysqli("127.0.0.1", "root", "pwd", "db");
$mysqli->query("CREATE TEMPORARY TABLE cats (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), tenant_id VARCHAR(255), UNIQUE KEY uniq_name (name))");

$mysqli->query("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_456') AS vals ON DUPLICATE KEY UPDATE name = vals.name, tenant_id = vals.tenant_id");

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: INSERT on table 'cats' sets 'tenant_id' to 'org_456' but tenant ID is 'org_123'.*
