# Track custom events

Use `\aikido\track()` to report events that only your application knows about, such as failed logins. [Playbooks](https://help.aikido.dev/zen-firewall/zen-features/playbooks) can act when an event occurs repeatedly, for example by blocking an IP after three failed logins in five minutes.

```php
if (!authenticate($_POST['email'], $_POST['password'])) {
    if (extension_loaded('aikido')) {
        \aikido\track('user.login_failed');
    }
    http_response_code(401);
    return;
}
```

After adding `\aikido\track()`, trigger the event at least once. It will then appear on the Playbooks page in the Aikido dashboard. From there, you can create a playbook and choose what should happen when the event occurs. Calling `\aikido\track()` by itself does not create a playbook or block anything.

Call `\aikido\track()` while handling an HTTP request. Zen associates the event with the request's IP address. Playbook counts are per IP, not across your whole app. If you call [`\aikido\set_user()`](./user.md) before `\aikido\track()`, Zen also includes the current user. `\aikido\set_user()` is optional. Events without a user are still tracked.

Event names can use any format. We recommend lowercase, dot-separated names such as `user.login_failed`.
