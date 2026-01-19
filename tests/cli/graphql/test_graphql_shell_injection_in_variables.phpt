--TEST--
Test shell injection detection in GraphQL variables

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1
REQUEST_URI=/graphql
HTTP_HOST=test.local
REQUEST_METHOD=POST


--POST_RAW--
Content-Type: application/json

{
  "query": "mutation ExportData($filename: String!) { export(filename: $filename) { success } }",
  "variables": {
    "filename": "data.txt; cat /etc/passwd"
  }
}

--FILE--
<?php
$input = file_get_contents('php://input');
$data = json_decode($input, true);

$filename = $data['variables']['filename'] ?? 'default.txt';

// Vulnerable shell command
$cmd = "echo 'Exporting to " . $filename . "'";
echo exec($cmd) . "\n";

?>

--EXPECTREGEX--
.*Detected GraphQL request.*
.*Aikido firewall has blocked a shell injection.*

