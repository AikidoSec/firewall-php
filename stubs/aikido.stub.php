<?php

/**
 * IDE stub for the Aikido Zen PHP extension.
 *
 * Copy this file into your own project, then set it up:
 *
 * PHPStorm:
 * No setup needed, as long as this file is in an indexed project folder.
 *
 * PHPStan:
 * Add it to `scanFiles` in phpstan.neon.
 *
 * Psalm:
 * Add it under `<stubs>` in psalm.xml. Not `<projectFiles>`: Psalm can't
 * resolve namespaced functions declared there.
 */

namespace {
    if (extension_loaded('aikido')) {
        return;
    }

    /**
     * Returned by `\aikido\should_block_request()`.
     */
    class AikidoBlockRequestStatus
    {
        /** Should this request be blocked? */
        public bool $block = false;

        /** Why it's blocked, e.g. "blocked" or "ratelimited". Empty if not blocked. */
        public string $type = '';

        /** What triggered it, e.g. "user", "ip" or "group". Empty if not blocked. */
        public string $trigger = '';

        /** Human-readable reason for the block. */
        public string $description = '';

        /** IP address of the request, if known. */
        public string $ip = '';

        /** User-Agent of the request, if known. */
        public string $user_agent = '';
    }

    /**
     * Returned by `\aikido\should_whitelist_request()`.
     */
    class AikidoWhitelistRequestStatus
    {
        /** Is this request whitelisted? */
        public bool $whitelisted = false;

        /** Type of whitelist that matched. Empty if not whitelisted. */
        public string $type = '';

        /** What triggered it, e.g. "ip". Empty if not whitelisted. */
        public string $trigger = '';

        /** Human-readable reason for the whitelist match. */
        public string $description = '';

        /** IP address of the request, if known. */
        public string $ip = '';
    }
}

namespace aikido {

    /**
     * Sets the current user. Used for rate limiting and shown in the dashboard.
     *
     * @param string $id User ID.
     * @param string $name Optional, shown in the dashboard.
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/user.md
     */
    function set_user(string $id, string $name = ''): bool
    {
    }

    /**
     * Turns on IDOR protection: checks that every SQL query filters on the given
     * tenant column, using the tenant id set with `set_tenant_id()`.
     *
     * @param string $tenantColumnName The column that says which tenant a row belongs to
     *   (e.g. "account_id", "organization_id", "team_id").
     * @param string[] $excludedTables Tables to skip, e.g. a shared "users" table with
     *   users from every tenant.
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/idor-protection.md
     */
    function enable_idor_protection(string $tenantColumnName, array $excludedTables = []): bool
    {
    }

    /**
     * Sets the tenant id for the current request/script. Every query is checked against
     * this tenant once `enable_idor_protection()` is on.
     *
     * @param string|int $tenantId
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/idor-protection.md
     */
    function set_tenant_id($tenantId): void
    {
    }

    /**
     * Returns the tenant id set with `set_tenant_id()`, or null if it was never set.
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/idor-protection.md
     */
    function get_tenant_id(): ?string
    {
    }

    /**
     * Runs the given callable with IDOR checks suspended, for queries that don't need
     * a tenant filter (e.g. an admin dashboard counting across all tenants).
     *
     * @template T
     * @param callable(): T $fn
     * @return T
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/idor-protection.md
     */
    function without_idor_protection(callable $fn)
    {
    }

    /**
     * Checks whether the current request should be blocked (user blocked,
     * or rate limit hit). Call after `set_user()`.
     *
     * Returns null outside of an HTTP request, e.g. in CLI.
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/should_block_request.md
     */
    function should_block_request(): ?\AikidoBlockRequestStatus
    {
    }

    /**
     * Checks whether the current request is whitelisted, based on IP.
     *
     * Returns null outside of an HTTP request, e.g. in CLI.
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/should_whitelist_request.md
     */
    function should_whitelist_request(): ?\AikidoWhitelistRequestStatus
    {
    }

    /**
     * Sets the Aikido token from PHP code, instead of an env var or ini setting.
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/token-setup-php.md
     */
    function set_token(string $token): bool
    {
    }

    /**
     * Groups the current request for rate limiting, e.g. by team or company,
     * instead of by individual user or IP.
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/user.md
     */
    function set_rate_limit_group(string $group): bool
    {
    }

    /**
     * Registers a custom route parameter, e.g. so "/posts/aikido-123" gets
     * recognized as "/posts/:tenant" instead of a literal path.
     *
     * @param string $param Parameter name, must match [a-zA-Z_]+.
     * @param string $regex Can contain placeholders like {digits} or {alpha}.
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/register-param-matcher.md
     */
    function register_param_matcher(string $param, string $regex): bool
    {
    }

    /**
     * Call at the start of each request when running in a persistent worker
     * (FrankenPHP worker mode, Laravel Octane).
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/frankenphp-worker.md
     */
    function worker_rinit(): bool
    {
    }

    /**
     * Call at the end of each request when running in a persistent worker
     * (FrankenPHP worker mode, Laravel Octane).
     *
     * @see https://github.com/AikidoSec/firewall-php/blob/main/docs/frankenphp-worker.md
     */
    function worker_rshutdown(): bool
    {
    }
}
