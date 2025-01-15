# Laravel Forge

There are two ways to install Aikido in Laravel forge.

- Portal: Use the UI and recipes functionality.
- SSH: Use SSH and standard package installation

## Portal

1. In Forge go to `[server_name] -> [site_name] -> Environment`, add the `AIKIDO_TOKEN` and `AIKIDO_BLOCKING` environment values and save. You can find their values in the Aikido platform.

2. In Forge go to "Recipes" and create a new recipe called "Install Aikido Firewall".

3. Based on the running OS, use the [Manual install](../README.md#Manual-install) commands to configure the new recipe and select "root" as user.

Example for Debian-based systems:
```
cd /tmp

# Install commands from the "Manual install" section below, based on your OS

curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.0.105/aikido-php-firewall.x86_64.deb
dpkg -i -E ./aikido-php-firewall.x86_64.deb

# Restarting the php services in order to load the Aikido PHP Firewall
for service in $(systemctl list-units | grep php | awk '{print $1}'); do
    sudo systemctl restart $service
done
```

4. Run the created recipes to install the Aikido PHP Firewall.

## SSH

1. In Forge go to `[server_name] -> [site_name] -> Environment`, add the `AIKIDO_TOKEN` and `AIKIDO_BLOCKING` environment values and save. You can find their values in the Aikido platform.

2. Use ssh to connect to the Forge server that you want to be protected by Aikido and, based on the running OS, execute the install commands from the [Manual install](../README.md#Manual-install) section.

3. Go to `[server_name] -> [site_name] -> Restart` and click `Restart PHP <version>`.
