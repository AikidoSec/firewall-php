--TEST--
Test PDOStatement::execute() method for SQL injection (GET url encoded + non valid UTF-8)

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--GET--
name=aa&age[]=3%27%29%3B%20DROP%20TABLE%20cats%3B%20--s%F0%28%8C%BC

--FILE--
<?php


try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
   
    $pdo->exec("CREATE TABLE IF NOT EXISTS cats (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        age INTEGER NOT NULL
    )");


    $stmt = $pdo->prepare("INSERT INTO cats (name, age) VALUES ( :name, '" . $_GET['age'][0] . "')");
    $stmt->execute([':name' => $_GET['name']]);

    echo "Query executed!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
