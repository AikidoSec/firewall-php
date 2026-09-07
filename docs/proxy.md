# Proxy settings

Zen uses `X-Forwarded-For` to determine the client's IP address by default. To use a different header, set `AIKIDO_CLIENT_IP_HEADER` to its HTTP header name, for example:

```sh
export AIKIDO_CLIENT_IP_HEADER=CF-Connecting-IP
```

Header names are case-insensitive. Use `CF-Connecting-IP`, without PHP's `HTTP_` prefix. If the setting is unset or empty, Zen uses `X-Forwarded-For`.

Zen selects the first valid, non-private IP address in the configured header's comma-separated list. If that header is missing or contains no usable IP, Zen falls back to `REMOTE_ADDR`. A custom header replaces `X-Forwarded-For`; Zen does not try both headers.

`AIKIDO_TRUST_PROXY` defaults to `true`. Set it to `false` to ignore IP headers and use `REMOTE_ADDR`, including when the app is exposed directly without a trusted reverse proxy. When proxy trust is enabled, configure your proxy to overwrite the selected header with the client's IP address.
