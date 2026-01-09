# syntax=docker/dockerfile:1.7
# CentOS Stream 9 test image with PHP built from source in NTS mode
# Used for testing the extension with standard PHP (non-thread-safe)

ARG BASE_IMAGE=quay.io/centos/centos:stream9
ARG PHP_VERSION=8.3
ARG PHP_SRC_REF=PHP-${PHP_VERSION}

FROM ${BASE_IMAGE} AS base
SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

ENV TZ=Etc/UTC \
    LC_ALL=C.UTF-8 \
    LANG=C.UTF-8 \
    PHP_VERSION=${PHP_VERSION}

RUN yum install -y yum-utils && \
    dnf config-manager --set-enabled crb || dnf config-manager --set-enabled powertools || true

# Install minimal tools needed for re2c build (replace curl-minimal with full curl)
RUN yum install -y xz tar gcc gcc-c++ make

ENV RE2C_VERSION=3.1
RUN curl -fsSL -o /tmp/re2c.tar.xz https://github.com/skvadrik/re2c/releases/download/${RE2C_VERSION}/re2c-${RE2C_VERSION}.tar.xz \
    && mkdir -p /tmp/re2c-src \
    && tar -xJf /tmp/re2c.tar.xz -C /tmp/re2c-src --strip-components=1 \
    && cd /tmp/re2c-src \
    && ./configure \
    && make -j"$(nproc)" \
    && make install \
    && rm -rf /tmp/re2c-src /tmp/re2c.tar.xz

# Install remaining build dependencies and tools
RUN yum install -y autoconf bison pkgconfig \
    libxml2-devel sqlite-devel libcurl-devel openssl-devel \
    libzip-devel oniguruma-devel libjpeg-turbo-devel libpng-devel libwebp-devel \
    libicu-devel readline-devel libxslt-devel \
    git wget \
    python3 python3-devel python3-pip \
    nginx httpd httpd-devel procps-ng mysql-server \
    cpio unzip nano lsof jq rpmdevtools sudo \
 && yum clean all

# Install mariadb-devel separately (may need different repo or skip if not critical)
RUN yum install -y mariadb-devel || yum install -y mariadb-connector-c-devel || echo "Warning: mariadb-devel not available, continuing without it"

# Install Go toolchain (architecture-aware)
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then \
        curl -O https://dl.google.com/go/go1.23.3.linux-amd64.tar.gz && \
        tar -C /usr/local -xzf go1.23.3.linux-amd64.tar.gz && \
        rm -f go1.23.3.linux-amd64.tar.gz; \
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then \
        curl -O https://dl.google.com/go/go1.23.3.linux-arm64.tar.gz && \
        tar -C /usr/local -xzf go1.23.3.linux-arm64.tar.gz && \
        rm -f go1.23.3.linux-arm64.tar.gz; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi
ENV PATH="/usr/local/go/bin:${PATH}"

# Install protoc and Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN ARCH=$(uname -m) && \
    if [ "$ARCH" = "x86_64" ] || [ "$ARCH" = "amd64" ]; then \
        PROTOC_ZIP=protoc-28.3-linux-x86_64.zip && \
        curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v28.3/$PROTOC_ZIP && \
        unzip -o $PROTOC_ZIP -d /usr/local bin/protoc && \
        unzip -o $PROTOC_ZIP -d /usr/local include/* && \
        rm -f $PROTOC_ZIP; \
    elif [ "$ARCH" = "aarch64" ] || [ "$ARCH" = "arm64" ]; then \
        PROTOC_ZIP=protoc-28.3-linux-aarch_64.zip && \
        curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v28.3/$PROTOC_ZIP && \
        unzip -o $PROTOC_ZIP -d /usr/local bin/protoc && \
        unzip -o $PROTOC_ZIP -d /usr/local include/* && \
        rm -f $PROTOC_ZIP; \
    else \
        echo "Unsupported architecture: $ARCH" && exit 1; \
    fi
ENV PATH="$HOME/go/bin:${PATH}"

# Fetch and build PHP from source with NTS
FROM base AS php-build
ARG PHP_SRC_REF
WORKDIR /usr/src
RUN git clone --depth 1 --branch "${PHP_SRC_REF}" https://github.com/php/php-src.git
WORKDIR /usr/src/php-src

RUN sed -i 's/\[\([0-9]\+\.[0-9]\+\.[0-9]\+\)-dev\]/[\1]/' configure.ac
RUN ./buildconf --force

# Patch openssl.c for OpenSSL compatibility (RSA_SSLV23_PADDING may not be available in newer OpenSSL)
RUN if [ -f ext/openssl/openssl.c ] && grep -q 'REGISTER_LONG_CONSTANT("OPENSSL_SSLV23_PADDING"' ext/openssl/openssl.c; then \
    awk '/REGISTER_LONG_CONSTANT\("OPENSSL_SSLV23_PADDING"/ { \
        indent = $0; \
        gsub(/[^[:space:]].*/, "", indent); \
        print indent "#ifdef RSA_SSLV23_PADDING"; \
        gsub(/^[[:space:]]*/, indent "    "); \
        print; \
        print indent "#endif"; \
        next \
    } \
    { print }' ext/openssl/openssl.c > ext/openssl/openssl.c.new && \
    mv ext/openssl/openssl.c.new ext/openssl/openssl.c; \
    fi || true

# Build PHP with NTS (no ZTS flags)
RUN ./configure \
      --prefix=/usr/local \
      --with-config-file-path=/usr/local/lib \
      --with-config-file-scan-dir=/usr/local/etc/php/conf.d \
      --enable-fpm \
      --enable-mbstring \
      --enable-pcntl \
      --with-extra-version="" \
      --with-curl \
      --with-mysqli \
      --with-openssl \
      --with-zlib \
      --with-zip \
      --with-apxs2=/usr/bin/apxs \
&& make -j"$(nproc)" \
&& make install \
&& strip /usr/local/bin/php /usr/local/sbin/php-fpm || true

# Final image with PHP and test infrastructure
FROM base AS final
COPY --from=php-build /usr/local /usr/local
# Copy libphp.so module (installed by apxs to Apache modules directory, not /usr/local)
# Note: /usr/lib64/ is the standard path for 64-bit libraries on both x86_64 and aarch64
# The architecture is determined by the binary itself (ELF header), not the directory path
# apxs installs to /usr/lib64/httpd/modules/ on CentOS/RHEL, or /usr/lib/httpd/modules/ on some distros
# We need to copy it from the build stage since COPY --from=php-build /usr/local only copies /usr/local
RUN mkdir -p /usr/lib64/httpd/modules /usr/lib/httpd/modules
# Copy from lib64 (standard on CentOS/RHEL for both x86_64 and aarch64)
# If libphp.so exists in the build stage, it will be copied; if not, COPY will fail gracefully
COPY --from=php-build /usr/lib64/httpd/modules/ /usr/lib64/httpd/modules/

RUN EXTENSION_DIR=$(php -i | grep "^extension_dir" | awk '{print $3}') && \
    if [ -z "$EXTENSION_DIR" ]; then \
        echo "Error: Could not determine extension_dir"; \
        exit 1; \
    fi && \
    mkdir -p "$EXTENSION_DIR" \
    echo "Created extension_dir: $EXTENSION_DIR"

# Verify NTS (ZTS should NOT be enabled)
RUN php -v | grep -v "ZTS" >/dev/null || (echo "ERROR: ZTS is enabled but should be NTS!" && exit 1) && \
    php -m | grep -E 'curl|mysqli' >/dev/null

ENV PATH="/usr/local/bin:${PATH}"

RUN ln -sf /usr/local/bin/php /usr/bin/php && \
    ln -sf /usr/local/sbin/php-fpm /usr/sbin/php-fpm || true

RUN     mkdir -p /etc/php-fpm.d && \
    mkdir -p /run/php-fpm && \
    mkdir -p /var/run && \
    mkdir -p /var/log/php-fpm && \
    mkdir -p /etc/httpd || true && \
    mkdir -p /usr/local/etc/php-fpm.d && \
    mkdir -p /usr/local/etc/php/conf.d && \

    ln -sf /usr/local/etc/php/conf.d /etc/php.d || true && \

    echo "[global]" > /usr/local/etc/php-fpm.conf && \
    echo "pid = /run/php-fpm/php-fpm.pid" >> /usr/local/etc/php-fpm.conf && \
    echo "error_log = /var/log/php-fpm/error.log" >> /usr/local/etc/php-fpm.conf && \
    echo "daemonize = yes" >> /usr/local/etc/php-fpm.conf && \
    echo "include=/usr/local/etc/php-fpm.d/*.conf" >> /usr/local/etc/php-fpm.conf && \
    echo "include=/etc/php-fpm.d/*.conf" >> /usr/local/etc/php-fpm.conf && \

    echo "[www]" > /usr/local/etc/php-fpm.d/www.conf && \
    echo "user = root" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "group = root" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "listen = 127.0.0.1:9000" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "listen.owner = root" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "listen.group = root" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "pm = dynamic" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "pm.max_children = 5" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "pm.start_servers = 2" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "pm.min_spare_servers = 1" >> /usr/local/etc/php-fpm.d/www.conf && \
    echo "pm.max_spare_servers = 3" >> /usr/local/etc/php-fpm.d/www.conf && \

    php-fpm -t -y /usr/local/etc/php-fpm.conf 2>&1 | grep -v "Nothing matches the include pattern" || true && \
    php-fpm -t -y /usr/local/etc/php-fpm.conf >/dev/null 2>&1 || \
    (PHP_MAJOR=$(php -r 'echo PHP_MAJOR_VERSION;') && \
     PHP_MINOR=$(php -r 'echo PHP_MINOR_VERSION;') && \
     echo "PHP-FPM config test failed for PHP ${PHP_MAJOR}.${PHP_MINOR}" && \
     echo "Config file contents:" && cat /usr/local/etc/php-fpm.conf && \
     echo "Pool config:" && cat /usr/local/etc/php-fpm.d/www.conf && \
     exit 1) && \

    ln -sf /usr/local/etc/php-fpm.conf /etc/php-fpm.conf && \
    ln -sf /usr/local/etc/php-fpm.d/www.conf /etc/php-fpm.d/www.conf || true

# Configure MySQL socket path for mysqli (so "localhost" connections work)
RUN mkdir -p /usr/local/etc/php/conf.d && \
    echo "mysqli.default_socket = /var/lib/mysql/mysql.sock" > /usr/local/etc/php/conf.d/mysql-socket.ini

RUN mkdir -p /etc/php/${PHP_VERSION}/apache2/conf.d && \
    echo "Created Apache mod_php directory: /etc/php/${PHP_VERSION}/apache2/conf.d"

# Configure Apache to load libphp.so module (if it exists)
RUN if [ -f /usr/lib64/httpd/modules/libphp.so ]; then \
    echo "# Load PHP module for Apache" > /etc/httpd/conf.modules.d/10-php.conf && \
    echo "LoadModule php_module modules/libphp.so" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "# Configure PHP file handling" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "<FilesMatch \.php$>" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "    SetHandler application/x-httpd-php" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "</FilesMatch>" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "# Directory index" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "DirectoryIndex index.php index.html" >> /etc/httpd/conf.modules.d/10-php.conf && \
    echo "Created Apache PHP module configuration at /etc/httpd/conf.modules.d/10-php.conf"; \
    else \
    echo "Warning: libphp.so not found, skipping Apache PHP module configuration"; \
    fi

# Python deps used by test harness
RUN python3 -m pip install --no-cache-dir --upgrade pip && \
    python3 -m pip install --no-cache-dir flask requests psutil

# Quality-of-life
WORKDIR /work
CMD ["bash"]


