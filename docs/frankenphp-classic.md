# FrankenPHP (Classic Mode)

1. Pass the Aikido environment variables in your `Caddyfile`

`Caddyfile`
```
example.com {
    root * /var/www/html/public

    php_server {
        env AIKIDO_TOKEN "AIK_RUNTIME_...."
        env AIKIDO_BLOCK "True"
    }

    file_server
}
```

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

2. Start FrankenPHP

`frankenphp run --config Caddyfile`
