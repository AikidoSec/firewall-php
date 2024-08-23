# Setting the current user

To set the current user, you can use the `aikido_set_user` php function.
We highly recommend to check that the aikido extension is loaded before making this function call, in order to prevent cases where the aikido extension was not yet deployed, or it was deployed after the server started.
Here's an example:

```php
// Get the user from your authentication middleware
// or wherever you store the user
if (extension_loaded('aikido')) {
    aikido_set_user("123", "John Doe");
}
});
```

Using `aikido_set_user` has the following benefits:
- The user ID is used for more accurate rate limiting (you can change IP addresses, but you can't change your user ID).
- Whenever attacks are detected, the user will be included in the report to Aikido.
- The dashboard will show all your users, where you can also block them.
- Passing the user's name is optional, but it can help you identify the user in the dashboard. You will be required to list Aikido Security as a subprocessor if you choose to share personal identifiable information (PII).
