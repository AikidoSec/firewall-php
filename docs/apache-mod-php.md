# Apache (mod_php)

1. Pass the Aikido environment variables to PHP from your Apache virtual host configuration (or .htaccess)

`/etc/apache2/sites-enabled/000-default.conf`
```
<VirtualHost *:80>
    ...
    
    SetEnv AIKIDO_TOKEN "AIK_RUNTIME_..."
    SetEnv AIKIDO_BLOCKING "False"

    ...

    <Directory "/var/www/html">
        ...
    </Directory>
</VirtualHost>
```

You can also use PassEnv if the environment is already configured at the system level.
```
<VirtualHost *:80>
    ...
    PassEnv AIKIDO_TOKEN
    PassEnv AIKIDO_BLOCKING
    ...
</VirtualHost>
```

2. Restart apache

(This command might differ on your operating system)

`service apache2 restart`
