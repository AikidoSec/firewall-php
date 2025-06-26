--TEST--
Test PDOStatement::execute() method for SQL injection (GET url encoded + non valid UTF-8)

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1


--POST_RAW--
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary7MA4YWxkTrZu0gW

------WebKitFormBoundary7MA4YWxkTrZu0gW
Content-Disposition: form-data; name="id"

id=1 AND sleep(3)-- =
------WebKitFormBoundary7MA4YWxkTrZu0gW--

--FILE--
<?php


try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
   
     $pdo->sqliteCreateFunction('sleep', function($seconds) {
        sleep($seconds);
        return $seconds;
    }, 1);
    
    $pdo->exec("CREATE TABLE IF NOT EXISTS cats (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age INTEGER NOT NULL
    )");

 
    $id = $_POST['id'];
    $query = "SELECT * FROM cats WHERE id = $id";


    $stmt = $pdo->prepare($query);
    $stmt->execute();

    echo "Query executed!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
