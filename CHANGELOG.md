# Changelog

## 2021-04-01: Released ContainerSSH 0.4.0

This version updates to the new ContainerSSH 0.4.0 release. 0.3.0 and 0.3.1 will remain supported for the time being.

## 2021-03-25: Upgraded OpenSSL to 1.1.1k

We have added an upgrade step to bump `libssl1.1` and `libcrypto1.1` to `1.1.1k-r0` to fix CVE-2021-3450 and CVE-2021-3449 [as announced by OpenSSL](https://www.openssl.org/news/vulnerabilities.html#y2021). This should not affect ContainerSSH since it has its own TLS implementation, but we still want to make sure.

## 2021-03-18: Fixed broken build process

In this version we fixed the previously-broken build process and images. Pleease see [the post mortem for details](https://containerssh.io/blog/2021/03/19/we-messed-up/).

## 2021-03-17: Initial version

This is the initial version for the new build process.
