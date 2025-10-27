# Troubleshooting

## Check logs for errors

`cat /var/log/aikido-*/*`

## Check if Aikido module has enabled

`php -i | grep "aikido support"`

Expected output: `aikido support => enabled`

## Switching PHP versions

Switching PHP versions after installing Zen is not currently supported. If you need to change your PHP version, you must uninstall and reinstall the firewall.

### Reinstall steps

1. **Uninstall the current package**

For Debian-based systems (Debian, Ubuntu):
```
dpkg --purge aikido-php-firewall
```

For Red Hat-based systems (RHEL, CentOS, Fedora):
```
rpm -e aikido-php-firewall
```

2. **Install Zen again**

Follow the installation instructions in the [README](../README.md#install) for your system architecture and distribution.
