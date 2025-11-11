# Laravel Forge

There are two ways to install Aikido in Laravel forge.

- Portal: Use the UI and recipes functionality.
- SSH: Use SSH and standard package installation

## Portal

1. In Forge go to your server -> `Settings` -> `Environment` and add the `AIKIDO_TOKEN` in the .env file. Optionally, you can set `AIKIDO_BLOCK` to 1 to enabling blocking mode for attacks.
![Forge Environment](./forge-environment.png)
You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

2. Go to "Commands" and run the following by replacing the sudo password with the one that Forge displays when the server is created:
```
curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.4.6/aikido-php-firewall.x86_64.deb && echo "YOUR_SUDO_PASSWORD_HERE" | sudo -S dpkg -i -E ./aikido-php-firewall.x86_64.deb && echo "YOUR_SUDO_PASSWORD_HERE" | sudo service php8.4-fpm restart
```
![Forge Commands](./forge-commands.png)

## SSH

1. In Forge go to `[server_name] -> [site_name] -> Environment`, add the `AIKIDO_TOKEN` and `AIKIDO_BLOCK` environment values and save. You can find their values in the Aikido platform.

2. Use ssh to connect to the Forge server that you want to be protected by Aikido and, based on the running OS, execute the install commands from the [Manual install](../README.md#Manual-install) section.

3. Run these bash lines to restart php-fpm:
```
# Restarting the php services in order to load the Aikido PHP Firewall
for service in $(systemctl list-units | grep php | awk '{print $1}'); do
    sudo systemctl restart $service
done
```
