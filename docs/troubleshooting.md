# Troubleshooting

## Review installation steps

Double-check your setup against the [installation guide](../README.md#installation).  
Make sure:
- The package installed correctly.  
- The firewall is initialized early in your app (before routes or handlers).  
- Your framework-specific integration (middleware, decorator, etc.) matches the example in the README.  
- You’re running a supported runtime version for your language.

## Check connection to Aikido

The firewall must be able to reach Aikido’s API endpoints.

Test from the same environment where your app runs and follow the instructions on this page: https://help.aikido.dev/zen-firewall/miscellaneous/outbound-network-connections-for-zen

## Check logs for errors

`cat /var/log/aikido-*/*`

## Check if Aikido module has enabled

`php -i | grep "aikido support"`

Expected output: `aikido support => enabled`

## Switching PHP versions

Switching PHP versions after installing Zen is currently not supported. If you need to change your PHP version, you must uninstall and reinstall the firewall:

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

## Contact support

If you still can’t resolve the issue:

- Use the in-app chat to reach our support team directly.
- Or create an issue on [GitHub](https://github.com/AikidoSec/firewall-php/issues) with details about your setup, framework, and logs.

Include as much context as possible (framework, logs, and how Aikido was added) so we can help you quickly.
