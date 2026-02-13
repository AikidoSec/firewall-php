--TEST--
Test that GraphQL-like requests to non-GraphQL routes are not detected

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
REQUEST_URI=/api/data
HTTP_HOST=test.local
REQUEST_METHOD=POST
AIKIDO_BLOCK=1

--POST_RAW--
Content-Type: application/json

{
  "query": "query { user(id: \"123\") { name } }"
}

--FILE--
<?php
$input = file_get_contents('php://input');
$data = json_decode($input, true);

echo "Not a GraphQL endpoint\n";

?>

--EXPECTREGEX--
(?s)\A(?!.*Detected GraphQL request).*?\z

