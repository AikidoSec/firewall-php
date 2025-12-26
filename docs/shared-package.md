Zen for PHP comes as a single package that needs to be installed on the system to be protected.

Prerequisites:
* Ensure you have sudo privileges on your system.
* Check that you have a supported PHP version installed (PHP version >= 7.2 and tested up to 8.5).
* Make sure to use the appropriate commands for your system or cloud provider.

#### For Red Hat-based Systems (RHEL, CentOS, Fedora)

##### x86_64
```
rpm -Uvh --oldpackage https://github.com/AikidoSec/firewall-php/releases/download/v1.4.11/aikido-php-firewall.x86_64.rpm
```

##### arm64 / aarch64
```
rpm -Uvh --oldpackage https://github.com/AikidoSec/firewall-php/releases/download/v1.4.11/aikido-php-firewall.aarch64.rpm
```

#### For Debian-based Systems (Debian, Ubuntu)

##### x86_64
```
curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.4.11/aikido-php-firewall.x86_64.deb
dpkg -i -E ./aikido-php-firewall.x86_64.deb
```

##### arm64 / aarch64
```
curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.4.11/aikido-php-firewall.aarch64.deb
dpkg -i -E ./aikido-php-firewall.aarch64.deb
```

We support Debian >= 11 and Ubuntu >= 20.04.
You can run on Debian 10, by doing this setup before install: [Debian10 setup](./docs/debian10.md)

#### Deployment setup
- [Caddy & PHP-FPM](./docs/caddy.md)
- [Apache mod_php](./docs/apache-mod-php.md)

#### Managed platforms
- [Laravel Forge](./docs/laravel-forge.md)
- [AWS Elastic beanstalk](./docs/aws-elastic-beanstalk.md)
- [Fly.io](./docs/fly-io.md)
