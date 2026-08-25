--TEST--
Test PDO pgsql query with a carriage-return comment injection

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST_RAW--
Content-Type: application/json
{
    "petname": "foo'),--\\r(version()||'\\n"
}

--FILE--
<?php
$pdo = new PDO(
    "pgsql:host=127.0.0.1;port=5432;dbname=zen_test",
    "postgres",
    "postgres"
);

$data = json_decode(file_get_contents('php://input'), true);

$pdo->exec("CREATE TEMPORARY TABLE cats_2 (petname TEXT)");
$pdo->exec("INSERT INTO cats_2 (petname) VALUES ('" . $data['petname'] . "');");

echo "Query executed\n";
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
