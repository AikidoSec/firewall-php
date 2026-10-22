<?php

if (isset($_GET["oom"])) {
    ini_set("memory_limit", "16M");
    $data = str_repeat("A", 64 * 1024 * 1024);
}

if (isset($_GET["timeout"])) {
    set_time_limit(1);
    while (true) {
    }
}

if (isset($_GET["exception"])) {
    throw new Exception("Something went wrong");
}

echo "Something";

?>
