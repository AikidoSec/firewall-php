# Should whitelist request

The `\aikido\should_whitelist_request` function allows the protected app to check whether the current request is whitelisted based on IP configuration. This can be used to skip custom security checks or apply special handling for requests coming from trusted or configured IPs.

## Function signature

```php
AikidoWhitelistRequestStatus \aikido\should_whitelist_request()
```

Returns an `AikidoWhitelistRequestStatus` object with the following properties:

| Property      | Type   | Description                                                         |
|---------------|--------|---------------------------------------------------------------------|
| `whitelisted` | bool   | Whether the request is whitelisted. Defaults to `false`.            |
| `type`        | string | The type of whitelist that matched. Empty string if not whitelisted.|
| `trigger`     | string | What triggered the whitelist (e.g., `"ip"`). Empty if not whitelisted. |
| `description` | string | A human-readable description of why the request is whitelisted.     |
| `ip`          | string | The IP address of the request. Empty if not whitelisted.            |

## Whitelist types

The function checks three conditions in order. The first match wins:

1. **`endpoint-allowlist`** — The endpoint has a route-level IP allowlist configured and the request IP is in it. This indicates that IP-based access control is active for this route.
2. **`bypassed`** — The request IP is in the global firewall bypass list.
3. **`allowlist`** — The request IP is found in the global allowed IP list (e.g., geo-location allow lists).

If none of the above conditions match, `whitelisted` is `false` and all other fields are empty strings.

## Example

```php
<?php

if (extension_loaded('aikido')) {
    $decision = \aikido\should_whitelist_request();

    if ($decision->whitelisted) {
        // The request is whitelisted — skip custom security checks
        // $decision->type contains the reason: "endpoint-allowlist", "bypassed", or "allowlist"
        // $decision->description has a human-readable explanation
    }
}
```
