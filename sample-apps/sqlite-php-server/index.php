<?php

if (isset($_GET['phpinfo'])) {
    phpinfo();
    exit();
}

$db = new SQLite3("cats.db");

$db->exec("CREATE TABLE IF NOT EXISTS cats (petname TEXT)");

if (isset($_GET['petname'])) {
    $petname = $_GET['petname'];
    $db->exec("INSERT INTO cats (petname) VALUES ('$petname')");
}

if (isset($_GET['clear'])) {
    $db->exec("DELETE FROM cats");
}

if (isset($_POST['cmd'])) {
    $cmd = $_POST['cmd'];
    $output = shell_exec($cmd);
}

if (isset($_GET['file'])) {
    $file = $_GET['file'];
    $fileContent = file_get_contents("/etc/" . $file);
}

$result = $db->query("SELECT petname FROM cats");
$cats = [];
while ($row = $result->fetchArray(SQLITE3_ASSOC)) {
    $cats[] = $row['petname'];
}
?>

<!DOCTYPE html>
<html lang="en">
<head>
    <title>Vulnerable Cats App</title>
</head>
<body>
    <p>All cats: <?= implode(", ", $cats) ?></p>
    <form action="/" method="GET">
        <label for="petname">Add a new cat</label>
        <input type="text" name="petname">
        <input type="submit" value="Add">
    </form>
    <a href="/?petname=Kitty'); DELETE FROM cats;--">Test SQL Injection</a> /
    <a href="/?clear=true">Clear table</a>

    <h2>Command Injection</h2>
    <form action="" method="POST">
        <label for="cmd">Run a command:</label>
        <input type="text" name="cmd">
        <input type="submit" value="Run">
    </form>
    <?php if (isset($output)) { echo "<pre>$output</pre>"; } ?>

    <h2>Path Traversal</h2>
    <form action="" method="GET">
        <label for="file">Read a file:</label>
        <input type="text" name="file">
        <input type="submit" value="Read">
    </form>
    <?php if (isset($fileContent)) { echo "<pre>$fileContent</pre>"; } ?>
</body>
</html>
