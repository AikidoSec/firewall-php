--TEST--
Test SQLite database operations

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--FILE--
<?php
try {
    $pdo = new PDO('sqlite::memory:');
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    $pdo->exec("CREATE TABLE IF NOT EXISTS users (
                id INTEGER PRIMARY KEY,
                name TEXT,
                email TEXT)");

    $pdo->exec("INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')");

    // Simulate user input
    $unsafeInput = "1' OR $$ IS NULL -- ";
    $_SERVER['HTTP_USER'] = $unsafeInput;

    // Vulnerable query
    $result = $pdo->query("SELECT * FROM users WHERE id = $unsafeInput");

    foreach ($result as $row) {
        echo "ID: " . $row['id'] . "\n";
        echo "Name: " . $row['name'] . "\n";
        echo "Email: " . $row['email'] . "\n\n";
    }
} catch (PDOException $e) {
    echo "Connection failed: " . $e->getMessage();
}

// Close the database connection
$pdo = null;

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
