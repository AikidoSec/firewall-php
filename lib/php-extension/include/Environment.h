#pragma once

#ifdef __MACH__
#define AIKIDO_INSTALL_DIR "/opt/homebrew/Cellar/aikido-php-firewall/" PHP_AIKIDO_VERSION "/aikido-" PHP_AIKIDO_VERSION 
#define AIKIDO_LOG_DIR "/opt/homebrew/var/log/aikido-" PHP_AIKIDO_VERSION
#define AIKIDO_RUN_DIR "/opt/homebrew/var/run/aikido-" PHP_AIKIDO_VERSION
#else
#define AIKIDO_INSTALL_DIR "/opt/aikido-" PHP_AIKIDO_VERSION
#define AIKIDO_LOG_DIR "/opt/aikido-" PHP_AIKIDO_VERSION
#define AIKIDO_RUN_DIR "/var/run/aikido-" PHP_AIKIDO_VERSION
#endif

void LoadEnvironment();
