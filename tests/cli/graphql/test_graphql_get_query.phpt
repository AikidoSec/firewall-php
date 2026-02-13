--TEST--
Test traversal attack detection in GraphQL GET query

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1
REQUEST_URI=/graphql
HTTP_HOST=test.local
REQUEST_METHOD=POST


--GET--
query=query+GetUsers+%7B%0A++users%28limit%3A+5%2C+offset%3A+0%2C+path%3A+%22..%2F..%2Fetc%2Fpasswd%22%29+%7B%0A++++id%0A++++name%0A++++email%0A++++posts+%7B%0A++++++id%0A++++++title%0A++++++createdAt%0A++++%7D%0A++%7D%0A%7D&variables=%7B%22limit%22%3A5%2C%22offset%22%3A0%7D
--FILE--
<?php
try {
  $file = '../../etc/passwd';
      
  file_get_contents($file);
} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>

--EXPECTREGEX--
.*Detected GraphQL request.*
.*Aikido firewall has blocked a path traversal attack: file_get_contents\(...\) originating from graphql.*


