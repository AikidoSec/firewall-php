--TEST--
Test SSRF detection in GraphQL with variables containing private IP

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1
REQUEST_URI=/graphql
HTTP_HOST=test.local
REQUEST_METHOD=POST


--POST_RAW--
Content-Type: application/json

{
  "query": "mutation SaveAsset($file: FileInput!) { save_asset(file: $file) { id } }",
  "variables": {
    "file": {
      "url": "http://127.0.0.1:8080/admin",
      "filename": "malicious.txt"
    }
  }
}

--FILE--
<?php
$input = file_get_contents('php://input');
$data = json_decode($input, true);

// Extract the URL from variables
$fileUrl = $data['variables']['file']['url'] ?? '';

echo "Attempting to fetch: " . $fileUrl . "\n";

// Attempt SSRF - should be blocked
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $fileUrl);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
curl_setopt($ch, CURLOPT_TIMEOUT, 5);
$result = curl_exec($ch);
curl_close($ch);

echo "File fetched successfully\n";

?>

--EXPECTREGEX--
.*Detected GraphQL request.*
.*Aikido firewall has blocked a server-side request forgery.*

