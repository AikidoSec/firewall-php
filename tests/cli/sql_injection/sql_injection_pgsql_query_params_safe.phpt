--TEST--
Test pg_query_params() keeps user input separate from the SQL query

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

$result = pg_query_params($db, 'SELECT $1::text', [$data['id']]);
echo "Value returned: " . pg_fetch_result($result, 0, 0) . "\n";

?>

--EXPECT--
Value returned: 1 OR 1=1 --
