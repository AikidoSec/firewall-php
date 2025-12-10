--TEST--
Test PHP shell execution functions

--ENV--
AIKIDO_LOG_LEVEL=INFO

--FILE--
<?php
// 1. exec()
$output = array();
$return_var = 0;
exec('echo "Hello from exec!"', $output, $return_var);
print_r($output);
echo "\n";

// 2. shell_exec()
$output = shell_exec('echo "Hello from shell_exec!"');
echo $output;
echo "\n";

// 3. system()
$return_var = 0;
system('echo "Hello from system!"', $return_var);
echo "\n";

// 4. passthru()
passthru('echo "Hello from passthru!"');
echo "\n";

// 5. popen()
$handle = popen('echo "Hello from popen!"', 'r');
while (!feof($handle)) {
    $buffer = fgets($handle);
    echo $buffer;
}
pclose($handle);
echo "\n";

// 6. proc_open()
$descriptorspec = array(
    0 => array("pipe", "r"),  // stdin is a pipe that the child will read from
    1 => array("pipe", "w"),  // stdout is a pipe that the child will write to
    2 => array("pipe", "w")   // stderr is a pipe that the child will write to
);

$process = proc_open('echo "Hello from proc_open!"', $descriptorspec, $pipes);

if (is_resource($process)) {
    while ($s = fgets($pipes[1])) {
        echo $s;
    }
    fclose($pipes[1]);
    proc_close($process);
}
echo "\n";
?>

--EXPECTREGEX--
\[AIKIDO\]\[INFO\]\[tid:\d+\] Got shell command: echo "Hello from exec!"
Array
\(
    \[0\] => Hello from exec!
\)

\[AIKIDO\]\[INFO\]\[tid:\d+\] Got shell command: echo "Hello from shell_exec!"
Hello from shell_exec!

\[AIKIDO\]\[INFO\]\[tid:\d+\] Got shell command: echo "Hello from system!"
Hello from system!

\[AIKIDO\]\[INFO\]\[tid:\d+\] Got shell command: echo "Hello from passthru!"
Hello from passthru!

\[AIKIDO\]\[INFO\]\[tid:\d+\] Got shell command: echo "Hello from popen!"
Hello from popen!

\[AIKIDO\]\[INFO\]\[tid:\d+\] Got shell command: echo "Hello from proc_open!"
Hello from proc_open!
