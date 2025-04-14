FROM php:8.3-apache

COPY 000-default.conf /etc/apache2/sites-enabled/000-default.conf

RUN apt update && \
    apt install unzip

RUN curl "https://wordpress.org/latest.zip" -o /root/latest.zip && \
    unzip /root/latest.zip -d /var/www/html/

RUN curl -L -O https://github.com/AikidoSec/firewall-php/releases/download/v1.0.109/aikido-php-firewall.aarch64.deb && \
    dpkg -i -E ./aikido-php-firewall.aarch64.deb
