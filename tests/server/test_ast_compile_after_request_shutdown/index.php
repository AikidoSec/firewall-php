<?php
// The require runs when the SAPI sends headers, after every module's RSHUTDOWN.
// Unique file per request so opcache cannot cache the compile.

$unique = sys_get_temp_dir() . '/late-header-helper-' . getmypid() . '-' . (int)(microtime(true) * 1e6) . '.php';
file_put_contents($unique, '<?php function late_header_helper_loaded(): bool { return true; }');

header_register_callback(function () use ($unique) {
    require $unique;
});

// Empty body: headers are sent at request shutdown.
http_response_code(204);
