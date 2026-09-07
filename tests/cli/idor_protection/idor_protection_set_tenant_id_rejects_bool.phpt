--TEST--
Test set_tenant_id() rejects boolean tenant IDs

--FILE--
<?php

\aikido\set_tenant_id(true);
var_dump(\aikido\get_tenant_id());

?>

--EXPECTF--
Warning: aikido\set_tenant_id(): tenantId must be a string or int in %s on line %d
NULL
