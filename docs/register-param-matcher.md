# Register custom route parameter matchers

Aikido exposes a new function, called `\aikido\register_param_matcher` that can be used to register custom route parameter matchers.
This allows you to define custom patterns for URL segments that should be treated as route parameters.

## Function signature

```php
bool \aikido\register_param_matcher(string $param, string $pattern)
```

- `$param`: The name of the parameter (e.g., "tenant", "org_id"). Must match `[a-zA-Z_]+`.
- `$pattern`: A pattern string that can contain placeholders like `{digits}` and `{alpha}`.
- Returns: `true` on success, `false` on failure.

## Supported placeholders

The pattern supports the following placeholders:

- `{digits}` - Matches one or more digits (`\d+`)
- `{alpha}` - Matches one or more alphabetic characters (`[a-zA-Z]+`)

## Pattern rules

1. The pattern must contain at least one placeholder (e.g., `{digits}` or `{alpha}`).
2. The pattern cannot contain slashes (`/`).
3. The pattern cannot contain consecutive similar placeholders (e.g., `{digits}{digits}`).

## Usage example

```php
<?php
if (extension_loaded('aikido')) {
    // Register a custom matcher for tenant IDs (e.g., "aikido-123")
    \aikido\register_param_matcher("tenant", "aikido-{digits}");
    
    // Register a custom matcher for org_ids (e.g., "aikido-foo-123-bar")
    \aikido\register_param_matcher("org_id", "aikido-{alpha}-{digits}-{alpha}");
}
?>
```

## When to use

This code needs to run with every request, so you can add it in a middleware, as exemplified [here](./should_block_request.md).
The backend will ignore duplicate registrations, so it's safe to call this function on every request.

## How it works

When Aikido processes URLs to build route patterns, it will check registered custom param matchers first.
If a URL segment matches a custom pattern, it will be replaced with `:param_name` in the route. For example:

- URL: `/posts/aikido-123` → Route: `/posts/:tenant` (if "tenant" matcher is registered)
- URL: `/blog/aikido-foo-123-bar` → Route: `/blog/:org_id` (if "org_id" matcher is registered)

If no custom matcher matches, Zen falls back to its default matchers (numbers, UUIDs, dates, etc.).
