# Docker container used for building Zen for PHP from source on Centos

FROM --platform=linux/amd64 centos:8

ARG PHP_VERSION=8.1

WORKDIR /etc/yum.repos.d/
RUN sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-*
RUN sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
RUN yum update -y
RUN yum install -y yum-utils
RUN yum install -y https://rpms.remirepo.net/enterprise/remi-release-8.4.rpm
RUN yum install -y httpd
RUN dnf --assumeyes module reset php
RUN dnf --assumeyes --nogpgcheck module install php:remi-${PHP_VERSION}
RUN dnf --assumeyes install php-pdo
RUN dnf --assumeyes install php-devel
RUN yum install -y mod_php
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
RUN yum install -y python3-devel
RUN yum install -y nginx
RUN pip3 install psutil flask requests --quiet --no-input
