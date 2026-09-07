--TEST--
Test IDOR protection remains enabled outside a suspended without_idor_protection() Fiber

--SKIPIF--
<?php
if (!class_exists('Fiber')) {
    die('skip Fibers require PHP 8.1 or newer');
}
?>

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php

\aikido\enable_idor_protection('tenant_id');
\aikido\set_tenant_id('org_123');

$pdo = new PDO('sqlite::memory:');
$pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
$pdo->exec('CREATE TABLE cats (id INTEGER PRIMARY KEY, name TEXT, tenant_id TEXT)');
$pdo->exec("INSERT INTO cats (name, tenant_id) VALUES ('Mittens', 'org_123')");

$fiber = new Fiber(function () {
    \aikido\without_idor_protection(function () {
        Fiber::suspend();
    });
});
$fiber->start();

$pdo->query('SELECT name FROM cats');

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Zen IDOR protection: query on table 'cats' is missing a filter on column 'tenant_id'.*
