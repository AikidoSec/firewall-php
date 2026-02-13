--TEST--
Test SQL injection detection in GraphQL variables

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1
REQUEST_URI=/graphql
HTTP_HOST=test.local
REQUEST_METHOD=POST


--POST_RAW--
Content-Type: application/json

{
  "query": "query GetUser($userId: String!) { user(id: \"1' OR '1'='1\") { name } }"
}

--FILE--
<?php
try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    
    $pdo->exec("CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL
    )");

    $userId = "1' OR '1'='1";
    
    // Vulnerable SQL query
    $query = "SELECT * FROM users WHERE id = '" . $userId . "'";
    $stmt = $pdo->prepare($query);
    $stmt->execute();

    echo "Query executed!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}

?>

--EXPECTREGEX--
.*Detected GraphQL request.*
.*Aikido firewall has blocked an SQL injection.*

