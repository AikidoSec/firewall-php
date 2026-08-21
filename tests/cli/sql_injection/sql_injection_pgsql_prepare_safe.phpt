--TEST--
Test pg_prepare() keeps user input separate from the prepared SQL query

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

pg_prepare($db, 'safe_statement', 'SELECT $1::text');
$result = pg_execute($db, 'safe_statement', [$data['id']]);
echo "Value returned: " . pg_fetch_result($result, 0, 0) . "\n";

?>

--EXPECT--
Value returned: 1 OR 1=1 --
