# AWS Elastic beanstalk

1. In your repo, create a new file in `.ebextensions/01_aikido_php_firewall.config` with the following content:
```
commands:
  aikido-php-firewall:
    command: "rpm -Uvh --oldpackage https://github.com/AikidoSec/firewall-php/releases/download/v1.3.0/aikido-php-firewall.x86_64.rpm"
    ignoreErrors: true

files:
  "/opt/elasticbeanstalk/tasks/bundlelogs.d/aikido-php-firewall.conf" :
    mode: "000755"
    owner: root
    group: root
    content: |
      /var/log/aikido-*/*.log

  "/opt/elasticbeanstalk/tasks/taillogs.d/aikido-php-firewall.conf" :
    mode: "000755"
    owner: root
    group: root
    content: |
      /var/log/aikido-*/*.log
```

2. Go to `AWS EB enviroment -> Configuration -> Updates, monitoring, and logging -> Edit`. add the `AIKIDO_TOKEN` and `AIKIDO_BLOCK` environment values and save.

You can get your token from the [Aikido Security Dashboard](https://help.aikido.dev/doc/creating-an-aikido-zen-firewall-token/doc6vRJNzC4u).
