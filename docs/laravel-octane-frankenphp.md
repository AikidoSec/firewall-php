# Laravel Octane + FrankenPHP

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

## Via Laravel Sail

1. Download Aikido installation package

`docker/version/Dockerfile`
```dockerfile
ARG AIKIDO_VERSION=1.5.2

RUN curl -L -o /tmp/aikido-php-firewall.deb \
    "https://github.com/AikidoSec/firewall-php/releases/download/v${AIKIDO_VERSION}/aikido-php-firewall.$(uname -m).deb" \
```

2. Pass the Aikido environment variables to FrankenPHP in Octane `Caddyfile` and configure the worker

`vendor/laravel/octane/src/Commands/stubs/Caddyfile`
```caddyfile
{$CADDY_SERVER_SERVER_NAME} {
    route {
        php_server {
            env AIKIDO_TOKEN "AIK_RUNTIME_...."
            env AIKIDO_BLOCK "True"
        }
    }
}
```

3. Call `\aikido\worker_rinit()` and `\aikido\worker_rshutdown()` in Octane worker script

Wrap your request handler with these calls to ensure Aikido processes each request.

`vendor/laravel/octane/bin/frankenphp-worker.php`
```php
<?php

// ... Octane bootstrap ...

$handleRequest = static function () use ($worker, $frankenPhpClient, $debugMode) {
    \aikido\worker_rinit();
    // Octane request handling logic
    \aikido\worker_rshutdown();
};

while ($requestCount < $maxRequests && frankenphp_handle_request($handleRequest)) {
    $requestCount++;
}
```

4. Install Aikido

Insert Aikido deb installation step in the end of the script.

`docker/version/start-container`
```
ls -sf /var/www/html/frankenphp /usr/bin/frankenphp
dpkg -i -E ./aikido-php-firewall.$(uname -m).deb
```

5. Start Sail

```bash
./vendor/bin/sail up --build
```

## Via Docker Image

1. Install Aikido in your `Dockerfile`

```dockerfile
FROM dunglas/frankenphp:php${PHP_VERSION}-bookworm

ARG AIKIDO_VERSION=1.5.2

RUN apt-get update && apt-get install -y curl \
    && curl -L -o /tmp/aikido-php-firewall.deb \
    "https://github.com/AikidoSec/firewall-php/releases/download/v${AIKIDO_VERSION}/aikido-php-firewall.$(uname -m).deb" \
    && dpkg -i /tmp/aikido-php-firewall.deb \
    && rm /tmp/aikido-php-firewall.deb \
    && apt-get clean

COPY . /app

ENTRYPOINT ["php", "artisan", "octane:frankenphp"]
```

2. Pass the Aikido environment variables to FrankenPHP in Octane `Caddyfile` and configure the worker

`vendor/laravel/octane/src/Commands/stubs/Caddyfile`
```caddyfile
{$CADDY_SERVER_SERVER_NAME} {
    route {
        php_server {
            env AIKIDO_TOKEN "AIK_RUNTIME_...."
            env AIKIDO_BLOCK "True"
        }
    }
}
```

3. Call `\aikido\worker_rinit()` and `\aikido\worker_rshutdown()` in Octane worker script

Wrap your request handler with these calls to ensure Aikido processes each request.

`vendor/laravel/octane/bin/frankenphp-worker.php`
```php
<?php

// ... Octane bootstrap ...

$handleRequest = static function () use ($worker, $frankenPhpClient, $debugMode) {
    \aikido\worker_rinit();
    // Octane request handling logic
    \aikido\worker_rshutdown();
};

while ($requestCount < $maxRequests && frankenphp_handle_request($handleRequest)) {
    $requestCount++;
}
```

4. Start the container

```bash
docker run -e AIKIDO_TOKEN="AIK_RUNTIME_..." your-app
```
