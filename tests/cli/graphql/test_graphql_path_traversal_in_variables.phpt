--TEST--
Test path traversal detection in GraphQL variables

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1
REQUEST_URI=/graphql
HTTP_HOST=test.local
REQUEST_METHOD=POST


--POST_RAW--
Content-Type: application/json

{
  "query": "query GetFile($path: String!) { file(path: $path) { content } }",
  "variables": {
    "path": "../../etc/passwd"
  }
}

--FILE--
<?php
$input = file_get_contents('php://input');
$data = json_decode($input, true);

$path = $data['variables']['path'] ?? 'default.txt';

// Vulnerable file access
$content = file_get_contents($path);
echo "File read successfully\n";

?>

--EXPECTREGEX--
.*Detected GraphQL request.*
.*Aikido firewall has blocked a path traversal attack.*

