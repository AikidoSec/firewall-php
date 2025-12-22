---
title: PHP Code
---

# Token setup from PHP code

Aikido exposes a new function, called `\aikido\set_token` that can be used to pass in the token to the Aikido Agent.

```php
<?php
if (extension_loaded('aikido')) {
    \aikido\set_token("your token here");
}
?>
```

This code needs to run with every request, so you can add it in a middleware, as exemplified [here](./should_block_request.md).
