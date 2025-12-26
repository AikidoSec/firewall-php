---
title: Caddy
eleventyNavigation:
  key: Caddy
  parent: Installation
---

# Caddy (PHP-FPM)

## 1. Pass the Aikido environment variables to PHP-FPM from your `Caddyfile`

`/etc/caddy/Caddyfile`
```diff-nginx
example.com {
    root * /var/www

    php_fastcgi unix//run/php/php-fpm.sock {
        ...
+        env AIKIDO_TOKEN "AIK_RUNTIME_...."
+        env AIKIDO_BLOCK "True"
        ...
    }
    file_server

    ...
}
```

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

## 2. Configure `PHP-FPM` to pass through the environment variables to PHP

`/etc/php/8.2/fpm/pool.d/www.conf`

```diff-ini
...
+ clear_env = no
+ env[AIKIDO_TOKEN] = $AIKIDO_TOKEN
+ env[AIKIDO_BLOCK] = $AIKIDO_BLOCK
```

## 3. Restart your Caddy and PHP-FPM services

(This command might differ on your operating system)

```bash
service caddy restart
service php8.2-fpm restart
```
