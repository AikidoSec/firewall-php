--TEST--
Test PHP shell injection (exec) in class 

--ENV--
AIKIDO_LOG_LEVEL=INFO
AIKIDO_BLOCK=1

--POST--
test=www.example`whoami`.com

--FILE--
<?php
class MyClass {
    public function test() {
        $output = [];
        $return_var = 0;

        exec('binary --domain www.example`whoami`.com', $output, $return_var);
        print_r($output);
    }
}

$instance = new MyClass();
$instance->test();
echo "\n";

?>

--EXPECTREGEX--
.*Fatal error: Uncaught Exception: Aikido firewall has blocked a shell injection.*