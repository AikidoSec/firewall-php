--TEST--
Test PDOStatement::execute() method for SQL injection (POST url encoded)

--ENV--
AIKIDO_LOG_LEVEL=DEBUG
AIKIDO_BLOCK=1

--POST--
name=Tom&age=3%27%29%3B%20DROP%20TABLE%20cats%3B%20--

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


    $stmt = $pdo->prepare("INSERT INTO cats (name, age) VALUES ( :name, '" . $_POST['age'] . "')");
    $stmt->execute([':name' => $_POST['name']]);

    echo "Cat " . $_POST['name'] . " added to the database successfully!";

} catch (PDOException $e) {
    echo "Error: " . $e->getMessage();
}
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
