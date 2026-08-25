--TEST--
Test an SQL injection detection does not skip IDOR protection in detection-only mode

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=0

--GET--
age=%27%7C%7Csqlite_version%28%29%7C%7C%27

--FILE--
<?php

\aikido\enable_idor_protection('tenant_id');
\aikido\set_tenant_id('org_123');

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec('CREATE TABLE cats (id INTEGER PRIMARY KEY, name TEXT, age TEXT, tenant_id TEXT)');
$pdo->exec("INSERT INTO cats (name, age, tenant_id) VALUES ('Mittens', '3', 'org_123')");

$pdo->query("SELECT name FROM cats WHERE age = '" . $_GET['age'] . "'");

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' is missing a filter on column 'tenant_id'.*
