--TEST--
Test SSRF detection in nested GraphQL mutation with private IP

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1
REQUEST_URI=/graphql
HTTP_HOST=test.local
REQUEST_METHOD=POST


--POST_RAW--
Content-Type: application/json

{"query":"mutation {\n    save_testVol_Asset(_file: { \n        url: \"http://0.0.0.0:80\"\n        filename: \"poc.txt\"\n    }) {\n        id\n    }\n}"}

--FILE--
<?php
$input = file_get_contents('php://input');
$data = json_decode($input, true);

// Extract the URL from the GraphQL query (simplified parsing)
preg_match('/url:\s*"([^"]+)"/', $data['query'], $matches);
$fileUrl = $matches[1] ?? '';

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

