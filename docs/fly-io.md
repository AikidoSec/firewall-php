# Fly.io (flyctl)

1. In your repo, run `fly launch`.

2. Add the desired environment variables, by running

- `fly secrets set AIKIDO_TOKEN=AIK_RUNTIME...`
- `fly secrets set AIKIDO_BLOCKING=false`

You can find their values in the Aikido platform.

3. Go to `./.fly/scripts` folder and create the `aikido.sh` file with the [Manual install](../README.md#Manual-install) commands:

```
#!/usr/bin/env bash
cd /tmp

curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.0.108/aikido-php-firewall.x86_64.deb
dpkg -i -E ./aikido-php-firewall.x86_64.deb
```

4. Run `fly deploy`.