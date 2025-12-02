FROM --platform=linux/amd64 centos:8

ARG PHP_VERSION=8.4
ARG PHP_FULL_VERSION=8.4.14
ARG FRANKENPHP_VERSION=1.9.1

WORKDIR /etc/yum.repos.d/
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
RUN yum update -y
RUN yum install -y yum-utils
RUN yum install -y https://rpms.remirepo.net/enterprise/remi-release-8.4.rpm
RUN yum install -y epel-release
RUN dnf config-manager --set-enabled powertools || dnf config-manager --set-enabled PowerTools || true
RUN yum install -y httpd
RUN yum install -y cpio
RUN yum install -y unzip
RUN yum install -y nano
RUN yum install -y lsof
RUN yum install -y jq
RUN yum install -y libcurl-devel
RUN curl -O https://dl.google.com/go/go1.23.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PROTOC_ZIP=protoc-28.3-linux-x86_64.zip
RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v28.3/$PROTOC_ZIP
RUN unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
RUN unzip -o $PROTOC_ZIP -d /usr/local include/*
RUN rm -f $PROTOC_ZIP
ENV PATH="$HOME/go/bin:${PATH}"
RUN yum install -y rpmdevtools
RUN yum install -y git
RUN yum install -y nginx
RUN yum install -y sudo
RUN yum install -y gcc gcc-c++ make
RUN dnf -y module enable python39 
RUN yum install -y python39 python39-devel
RUN ln -sf /usr/bin/python3.9 /usr/bin/python3 && ln -sf /usr/bin/python3.9 /usr/bin/python
RUN pip3 install psutil flask requests --quiet --no-input

# Checkout PHP sources at the requested full version tag/branch (e.g., PHP-8.4.14)
RUN git clone --depth 1 --branch PHP-${PHP_FULL_VERSION} https://github.com/php/php-src.git /usr/local/src/php-src

## Build and install a recent re2c required by PHP build system
ENV RE2C_VERSION=3.1
RUN curl -fsSL -o /tmp/re2c.tar.xz https://github.com/skvadrik/re2c/releases/download/${RE2C_VERSION}/re2c-${RE2C_VERSION}.tar.xz \
    && mkdir -p /tmp/re2c-src \
    && tar -xJf /tmp/re2c.tar.xz -C /tmp/re2c-src --strip-components=1 \
    && cd /tmp/re2c-src \
    && ./configure \
    && make -j"$(nproc)" \
    && make install

RUN yum install -y \
      autoconf automake libtool bison pkgconfig \
      libxml2-devel \
      oniguruma-devel \
      libicu-devel \
      mariadb-devel \
      openssl-devel \
      zlib-devel \
      libzip-devel \
      sqlite-devel \
    && cd /usr/local/src/php-src \
    && sed -i 's/\[\([0-9]\+\.[0-9]\+\.[0-9]\+\)-dev\]/[\1]/' configure.ac \
    && ./buildconf --force \
    && CFLAGS="$CFLAGS -fPIE -fPIC" LDFLAGS="$LDFLAGS -pie" ./configure \
         --enable-embed \
         --enable-zts \
         --enable-pdo \
         --disable-zend-signals \
         --enable-zend-max-execution-timers \
         --with-extra-version="" \
         --with-config-file-scan-dir=/etc/php.d \
         --enable-mbstring \
         --enable-pcntl \
         --with-curl \
         --with-mysqli \
         --with-openssl \
         --with-zlib \
         --with-zip \
         --with-sqlite3 \
         --with-pdo-sqlite \
    && make -j"$(nproc)" \
    && make install 

RUN mkdir -p /etc/php.d

RUN ln -sf /usr/local/bin/php /usr/bin/php

# Install FrankenPHP from RPM
RUN FRANKENPHP_RPM_URL="https://github.com/php/frankenphp/releases/download/v${FRANKENPHP_VERSION}/frankenphp-${FRANKENPHP_VERSION}-1.x86_64.rpm" \
    && curl -fsSL -L -o /tmp/frankenphp.rpm "$FRANKENPHP_RPM_URL" \
    && yum install -y /tmp/frankenphp.rpm \
    && rm -f /tmp/frankenphp.rpm

 
