# Support matrix

Zen for PHP supports PHP 7.2 through PHP 8.5 on x86_64 and aarch64 systems.

## Platforms

| Platform | Package |
| --- | --- |
| Debian 11 or newer | DEB |
| Ubuntu 20.04 or newer | DEB |
| RHEL, CentOS, and Fedora | RPM |
| Alpine Linux 3.20 or newer | APK |

## PHP runtimes

| Runtime | PHP versions | ABI |
| --- | --- | --- |
| PHP CLI and built-in server | 7.2-8.5 | NTS and ZTS |
| PHP-FPM | 7.2-8.5 | NTS and ZTS |
| Apache mod_php | 7.2-8.5 | NTS |
| FrankenPHP classic and worker | 8.2-8.5 | ZTS |

The package automatically selects the extension matching the installed PHP version and ABI. If you change PHP versions after installing Zen, reinstall the package as described in [Switching PHP versions](./troubleshooting.md#switching-php-versions).
