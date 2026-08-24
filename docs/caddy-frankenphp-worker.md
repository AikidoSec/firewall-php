# FrankenPHP (Worker Mode)

1. Pass the Aikido environment variables to FrankenPHP in your `Caddyfile` and configure the worker

`Caddyfile`
```
example.com {
    root * /var/www/html/public

    php_server {
        env AIKIDO_TOKEN "AIK_RUNTIME_...."
        env AIKIDO_BLOCK "True"
        worker {
            file /var/www/html/public/index.php
            num 4
        }
    }

    file_server
}
```

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

2. Call `\aikido\worker_rinit()` and `\aikido\worker_rshutdown()` in your worker script

Wrap your request handler with these calls to ensure Aikido processes each request.

`public/index.php`
```php
<?php

require __DIR__ . '/../vendor/autoload.php';
$app = require_once __DIR__ . '/../bootstrap/app.php';

while (frankenphp_handle_request(function () use ($app) {
    \aikido\worker_rinit();
    // Your application logic
    \aikido\worker_rshutdown();
})) {
    // keep looping
}
```

> [!NOTE]
> If you use [IDOR protection](idor-protection.md), call `\aikido\enable_idor_protection(...)` inside the handler, right after `worker_rinit()`, not in the one-time bootstrap code above the loop. `worker_rinit()` resets the IDOR protection config on every request, so a call made only during bootstrap won't survive past the first request:
>
> ```php
> while (frankenphp_handle_request(function () use ($app) {
>     \aikido\worker_rinit();
>     \aikido\enable_idor_protection("tenant_id", ["users"]);
>     // Your application logic, including your auth middleware, which is
>     // where you call \aikido\set_tenant_id() once the current user is known
>     \aikido\worker_rshutdown();
> })) {
>     // keep looping
> }
> ```

3. Start FrankenPHP

`frankenphp run --config Caddyfile`
