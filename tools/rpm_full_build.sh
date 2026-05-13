rpm -e aikido-php-firewall || true
set -e
./tools/build.sh && ./tools/rpm_build.sh && ./tools/rpm_install.sh
