---
title: Apache (mod_php)
---

# Apache (mod_php)

Pass the Aikido environment variables to PHP from your Apache virtual host configuration (or .htaccess)

`/etc/apache2/sites-enabled/000-default.conf`
```diff-apache
 <VirtualHost *:80>
    ...
    
+    SetEnv AIKIDO_TOKEN "AIK_RUNTIME_..."
+    SetEnv AIKIDO_BLOCK "False"

    ...

    <Directory "/var/www/html">
        ...
    </Directory>
 </VirtualHost>
```

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

You can also use PassEnv if the environment is already configured at the system level.
```diff-apache
 <VirtualHost *:80>
    ...
+    PassEnv AIKIDO_TOKEN
+    PassEnv AIKIDO_BLOCK
    ...
 </VirtualHost>
```

Restart apache

(This command might differ on your operating system)

```bash
service apache2 restart
```
