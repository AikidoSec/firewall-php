--TEST--
Test IDOR protection blocks a wrong pg_query_params() tenant value

--SKIPIF--
<?php
if (!extension_loaded('pgsql')) {
    die('skip pgsql not available');
}
if (!@pg_connect('host=127.0.0.1 port=5432 dbname=zen_test user=postgres password=postgres')) {
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

$db = pg_connect('host=127.0.0.1 port=5432 dbname=zen_test user=postgres password=postgres');
pg_query($db, 'CREATE TEMPORARY TABLE cats (id SERIAL PRIMARY KEY, name TEXT, tenant_id TEXT)');
pg_query($db, "INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

pg_query_params($db, 'SELECT name FROM cats WHERE tenant_id = $1', ['org_456']);

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' filters 'tenant_id' with value 'org_456' but tenant ID is 'org_123'.*
