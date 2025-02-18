# Docker container used for building Zen for PHP from source on Centos

FROM ubuntu:20.04

ARG PHP_VERSION=8.1

RUN apt-get update
RUN apt install software-properties-common -y
RUN add-apt-repository ppa:ondrej/php -y
RUN apt update
RUN apt install php${PHP_VERSION} php${PHP_VERSION}-cli php${PHP_VERSION}-cgi php${PHP_VERSION}-fpm php${PHP_VERSION}-dev php${PHP_VERSION}-curl php${PHP_VERSION}-sqlite3 -y
RUN apt-get install -y jq wget autoconf bison re2c libxml2-dev libssl-dev libcurl4-gnutls-dev protobuf-compiler protobuf-compiler-grpc git alien jq
RUN wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="${HOME}/go"
ENV PATH="${PATH}:${GOPATH}/bin"
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /firewall-php

COPY . /firewall-php

RUN chmod +x /firewall-php/tools/sample_apps_build.sh && /firewall-php/tools/sample_apps_build.sh

