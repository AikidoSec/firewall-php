# Caddy (PHP-FPM)

1. Pass the Aikido environment variables to PHP-FPM from your `Caddyfile`

`/etc/caddy/Caddyfile`
```
example.com {
    root * /var/www

    php_fastcgi unix//run/php/php-fpm.sock {
        ...
        env AIKIDO_TOKEN "AIK_RUNTIME_...."
        env AIKIDO_BLOCKING "False"
        ...
    }
    file_server

    ...
}
```

2. Configure `PHP-FPM` to pass through the environment variables to PHP

`/etc/php/8.2/fpm/pool.d/www.conf`

```
...
clear_env = no
env[AIKIDO_TOKEN] = $AIKIDO_TOKEN
env[AIKIDO_BLOCKING] = $AIKIDO_BLOCKING
```

3. Restart your Caddy and PHP-FPM services

(This command might differ on your operating system)

`service caddy restart`

`service php8.2-fpm restart`
