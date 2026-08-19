--TEST--
Test pg_send_query() with a benign numeric id is not blocked

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST_RAW--
Content-Type: application/json
{
    "id": "1"
}

--FILE--
<?php
$db = pg_connect("host=127.0.0.1 port=5432 dbname=zen_test user=postgres password=postgres");
if (!$db) { die("connect failed\n"); }

pg_query($db, "DROP TABLE IF EXISTS users");
pg_query($db, "CREATE TABLE users (id serial primary key, name text, secret text)");
pg_query($db, "INSERT INTO users (name, secret) VALUES ('alice', 's3cr3t')");

$data = json_decode(file_get_contents('php://input'), true);

pg_send_query($db, "SELECT * FROM users WHERE id = " . $data['id']);
$result = pg_get_result($db);
echo "Rows returned: " . pg_num_rows($result) . "\n";

?>

--EXPECT--
Rows returned: 1
