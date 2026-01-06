# Fly.io (flyctl)

## Installation Steps

### 1. Create a Fly.io Application

If you haven't already, initialize your Fly.io application:

```bash
fly launch
```

### 2. Set Environment Variables

Add the required Aikido environment variables to your Fly.io application:

```bash
fly secrets set AIKIDO_TOKEN=AIK_RUNTIME_...  # Replace with your actual token
fly secrets set AIKIDO_BLOCK=false         # Set to "true" to enable blocking mode
```

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).

### 3. Create Installation Script

Create a script to install the Aikido PHP Firewall during deployment:

1. Go to the `./.fly/scripts` folder in your repository
2. Create an `aikido.sh` file with the following content:

```bash
#!/usr/bin/env bash
cd /tmp

curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.4.12/aikido-php-firewall.x86_64.deb
dpkg -i -E ./aikido-php-firewall.x86_64.deb
```

3. Make the script executable:

```bash
chmod +x ./.fly/scripts/aikido.sh
```

### 4. Deploy Your Application

Deploy your application to Fly.io:

```bash
fly deploy
```

### 5. Verify Installation

You can verify the installation by connecting to your Fly.io instance and checking if the Aikido module is loaded:

```bash
fly ssh console
php -i | grep "aikido support"
```

Expected output: `aikido support => enabled`

If you encounter any issues, check the logs:

```bash
fly ssh console
cat /var/log/aikido-*/*
```

## Getting Your Aikido Token

For instructions on creating a token, see our [help documentation](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).
