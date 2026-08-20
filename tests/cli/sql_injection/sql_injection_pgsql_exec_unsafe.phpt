--TEST--
Test pg_exec() with a classic SQL injection payload built via string concatenation

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST_RAW--
Content-Type: application/json
{
    "id": "1 OR 1=1 --"
}

--FILE--
<?php
$db = pg_connect("host=127.0.0.1 port=5432 dbname=zen_test user=postgres password=postgres");
if (!$db) { die("connect failed\n"); }

$data = json_decode(file_get_contents('php://input'), true);

$result = pg_exec($db, "SELECT * FROM (VALUES (1), (2)) AS users(id) WHERE id = " . $data['id']);
echo "Rows returned: " . pg_num_rows($result) . "\n";

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
