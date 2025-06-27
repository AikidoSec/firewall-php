--TEST--
Test PDOStatement::execute() method for SQL injection (x-www-form-urlencoded)

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--POST--

id=1 AND sleep(3)-- =
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
