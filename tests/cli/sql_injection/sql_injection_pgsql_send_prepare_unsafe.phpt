--TEST--
Test pg_send_prepare() with a classic SQL injection payload in the prepared SQL query

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

pg_send_prepare($db, 'unsafe_async_statement', "SELECT * FROM (VALUES (1), (2)) AS users(id) WHERE id = " . $data['id']);
pg_get_result($db);
echo "Query was prepared\n";

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
