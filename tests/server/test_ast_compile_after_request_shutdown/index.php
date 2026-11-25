<?php
// Installs a session save handler whose write() calls
// eval(). PHP runs session RSHUTDOWN (and thus this callback) after Zen's
// RSHUTDOWN, so the eval compiles once Zen's AST cleanup table is freed.
class LateCompileSaveHandler implements SessionHandlerInterface {
    public function open($path, $name): bool { return true; }
    public function close(): bool { return true; }
    public function read($id): string { return ''; }
    public function write($id, $data): bool {
        // Compile after Zen's RSHUTDOWN: on a vulnerable build this poisons
        // globalAstToClean; the next compile in this worker SIGSEGVs.
        eval('return 43;');
        return true;
    }
    public function destroy($id): bool { return true; }
    public function gc($max_lifetime): int { return 0; }
}

session_set_save_handler(new LateCompileSaveHandler());
session_start();
$_SESSION['value'] = 1;

// Also compile during the request itself, so this worker has live AST
// instrumentation state going into shutdown.
$normal = eval('return 42;');

echo "value=" . $normal . PHP_EOL;
