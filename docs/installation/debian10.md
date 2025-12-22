# Debian10

1. In order to install the Aikido PHP firewall on Debian 10, you would need first to install gcc version 10:

```
wget https://mirrors.edge.kernel.org/ubuntu/pool/main/g/gcc-10/gcc-10-base_10-20200411-0ubuntu1_amd64.deb && dpkg -i gcc-10-base_10-20200411-0ubuntu1_amd64.deb
wget https://mirrors.edge.kernel.org/ubuntu/pool/main/g/gcc-10/libgcc-s1_10-20200411-0ubuntu1_amd64.deb && dpkg -i libgcc-s1_10-20200411-0ubuntu1_amd64.deb
```

Be aware that manually installing packages from a newer release can cause dependency conflicts. Ensure that your system's other packages remain compatible and test in staging environment before releasing.
