--TEST--
Test MySQLi database operations

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST_RAW--
Content-Type: application/json
{
    "test": "1 OR 1=1"
}

--FILE--
<?php
// Create table
mysqli_query(null, "CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(255),
    email VARCHAR(255)
)");

// Insert test data
mysqli_query(null, "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')");

// Simulate user input
$unsafeInput = "1 OR 1=1";

// Vulnerable query
mysqli_query(null, "SELECT * FROM users WHERE id = $unsafeInput");

if ($result) {
    while ($row = $result->fetch_assoc()) {
        echo "ID: " . $row['id'] . "\n";
        echo "Name: " . $row['name'] . "\n";
        echo "Email: " . $row['email'] . "\n\n";
    }
} else {
    echo $mysqli->error;
}

// Close the connection
$mysqli_close();
?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked an SQL injection.*
