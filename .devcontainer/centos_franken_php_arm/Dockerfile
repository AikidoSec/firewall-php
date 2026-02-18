FROM --platform=linux/arm64 centos:8

ARG PHP_VERSION=8.4
ARG PHP_FULL_VERSION=8.4.14
ARG FRANKENPHP_VERSION=1.9.1

WORKDIR /etc/yum.repos.d/
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* \
    && sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-* \
    && yum update -y \
    && yum install -y yum-utils \
    && yum install -y https://rpms.remirepo.net/enterprise/remi-release-8.4.rpm \
    && yum install -y epel-release \
    && dnf config-manager --set-enabled powertools || dnf config-manager --set-enabled PowerTools || true \
    && yum install -y httpd cpio unzip nano lsof jq libcurl-devel rpmdevtools git nginx sudo gcc gcc-c++ make \
    && yum clean all && rm -rf /var/cache/yum

# Go toolchain (arm64) and protoc (aarch_64) similar to centos_arm
RUN curl -O https://dl.google.com/go/go1.23.3.linux-arm64.tar.gz \
    && tar -C /usr/local -xzf go1.23.3.linux-arm64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PROTOC_ZIP=protoc-30.2-linux-aarch_64.zip
RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v30.2/$PROTOC_ZIP \
    && unzip -o $PROTOC_ZIP -d /usr/local bin/protoc \
    && unzip -o $PROTOC_ZIP -d /usr/local include/* \
    && rm -f $PROTOC_ZIP
ENV PATH="$HOME/go/bin:${PATH}"

# Python 3.9 for PHP build helpers
RUN dnf -y module enable python39 \
    && yum install -y python39 python39-devel \
    && ln -sf /usr/bin/python3.9 /usr/bin/python3 && ln -sf /usr/bin/python3.9 /usr/bin/python

RUN pip3 install psutil flask requests --quiet --no-input

# Fetch PHP sources
RUN git clone --depth 1 --branch PHP-${PHP_FULL_VERSION} https://github.com/php/php-src.git /usr/local/src/php-src

# Build re2c from source (arm64)
ENV RE2C_VERSION=3.1
RUN curl -fsSL -o /tmp/re2c.tar.xz https://github.com/skvadrik/re2c/releases/download/${RE2C_VERSION}/re2c-${RE2C_VERSION}.tar.xz \
    && mkdir -p /tmp/re2c-src \
    && tar -xJf /tmp/re2c.tar.xz -C /tmp/re2c-src --strip-components=1 \
    && cd /tmp/re2c-src \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && rm -rf /tmp/re2c.tar.xz /tmp/re2c-src

# Build latest GDB from release tarball (more stable than git)
ENV GDB_VERSION=14.2
RUN yum install -y texinfo mpfr-devel gmp-devel ncurses-devel readline-devel zlib-devel expat-devel \
    libmpc-devel gettext-devel \
    && curl -fsSL -o /tmp/gdb-${GDB_VERSION}.tar.xz https://ftp.gnu.org/gnu/gdb/gdb-${GDB_VERSION}.tar.xz \
    && mkdir -p /tmp/gdb-src \
    && tar -xJf /tmp/gdb-${GDB_VERSION}.tar.xz -C /tmp/gdb-src --strip-components=1 \
    && cd /tmp/gdb-src \
    && ./configure --prefix=/usr/local \
                   --with-python=/usr/bin/python3.9 \
                   --enable-tui \
                   --with-readline \
    && make -j"$(nproc)" \
    && make install \
    && echo 'set auto-load safe-path /' > /root/.gdbinit 

RUN yum install -y \
    autoconf automake libtool bison pkgconfig \
    libxml2-devel \
    oniguruma-devel \
    libicu-devel \
    mariadb-devel \
    openssl-devel \
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

# Try to install FrankenPHP for arm64 (optional)
RUN FRANKENPHP_ARCH=aarch64 \
    && FRANKENPHP_RPM_URL="https://github.com/php/frankenphp/releases/download/v${FRANKENPHP_VERSION}/frankenphp-${FRANKENPHP_VERSION}-1.${FRANKENPHP_ARCH}.rpm" \
    && curl -fsSL -L -o /tmp/frankenphp.rpm "$FRANKENPHP_RPM_URL" \
    && yum install -y /tmp/frankenphp.rpm || true \
    && rm -f /tmp/frankenphp.rpm || true

# Final PATH to include phpize/php-config if installed
ENV PATH="/usr/local/bin:/usr/local/sbin:${PATH}"


