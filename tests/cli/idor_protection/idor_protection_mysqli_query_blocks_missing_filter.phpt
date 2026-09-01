--TEST--
Test IDOR protection blocks mysqli::query() SELECT without a tenant filter

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
$mysqli->query("CREATE TEMPORARY TABLE cats (id INT AUTO_INCREMENT PRIMARY KEY, name VARCHAR(255), tenant_id VARCHAR(255))");
$mysqli->query("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

$mysqli->query("SELECT name FROM cats");

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' is missing a filter on column 'tenant_id'.*
