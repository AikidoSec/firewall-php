# IDOR Protection

IDOR (Insecure Direct Object Reference) is when one account can read or change another account's data because a query forgot to filter by account.

If your app has accounts (or organizations, workspaces, teams, ...) and keeps each account's data apart with a column like `tenant_id`, Zen makes sure every SQL query filters on the right tenant. It checks queries at runtime and throws an error when one is missing the filter or uses the wrong tenant ID. For example:

- A `SELECT` without the tenant filter, so one account can read another's orders
- An `UPDATE` or `DELETE` without the filter, so one account can change another's data
- An `INSERT` that leaves out the tenant column, creating rows that belong to nobody or the wrong account

Most of the time this is a bug you'll catch in development or testing. Sometimes it's an actual attack. Either way, Zen blocks the query before it runs. See [IDOR vulnerability explained](https://www.aikido.dev/blog/idor-vulnerability-explained) for more.

> [!NOTE]
> IDOR violations always throw, even when `AIKIDO_BLOCK` is off (unlike SQL injection, which Zen blocks or just reports depending on your mode). This works even without a token configured.
>
> The exception is a plain `\Exception` whose message starts with `Zen IDOR protection:`. Note that a broad `catch (\Exception $e)` around your data layer will swallow it like any other error.

## Setup

### 1. Turn it on

```php
if (extension_loaded('aikido')) {
    \aikido\enable_idor_protection("tenant_id", ["users"]);
}
```

- First argument — the column that says which tenant a row belongs to (e.g. `account_id`, `organization_id`, `team_id`).
- Second argument — tables to skip, because their rows aren't tied to one tenant (e.g. a shared `users` table with users from every account).

### 2. Set the tenant ID

Call `set_tenant_id` early in the request, usually in middleware once you know who the user is. Zen then checks every query in that request against this tenant:

```php
// Get the tenant ID from your auth layer
\aikido\set_tenant_id($user->organizationId);
```

The tenant stays set for the rest of the request. If a query runs and `set_tenant_id` was never called, Zen throws.

That's everything you need for code that runs inside requests. The sections below are optional.

## Supported databases

- SQLite, MySQL and PostgreSQL via `PDO` (`PDO::query`, `PDO::exec`, `PDOStatement::execute`). Values are resolved whether you pass them to `execute($params)` or bind them beforehand with `bindValue()` / `bindParam()`, with either named (`:tenant_id`) or positional (`?`) placeholders.
- `mysqli` for queries with the values written directly in the query string (`mysqli_query`, `mysqli_real_query`).

## Advanced options

<details>
<summary>Joins</summary>

Each table needs its own tenant filter. In an `INNER JOIN`, a filter in the `ON` clause counts:

```sql
SELECT t.id FROM tickets t
JOIN comments c ON c.ticket_id = t.id AND c.tenant_id = t.tenant_id
WHERE t.tenant_id = ?
```

`LEFT`, `RIGHT` and `FULL` joins are different: Zen ignores their `ON` clause, so give the joined table its own filter instead:

```sql
SELECT t.id FROM tickets t
LEFT JOIN comments c ON c.ticket_id = t.id AND c.tenant_id = ?
WHERE t.tenant_id = ?
```

</details>

<details>
<summary>Background work: set the tenant ID for every iteration</summary>

`set_tenant_id` takes a string or an int and compares as a string, so `set_tenant_id(123)` matches `123` or `'123'`. Bad input (`null`, `''`, an array, an object) triggers an `E_WARNING` and leaves the tenant unchanged. You can't unset it once set.

In a long-lived process like a queue worker, the tenant carries over between jobs. Call `set_tenant_id` at the top of every iteration:

```php
foreach ($jobs as $job) {
    \aikido\set_tenant_id($job['tenant_id']);
    processJob($job);
}
```

Bad input here means the job silently runs under the previous job's tenant. This doesn't affect php-fpm or mod_php, since the tenant resets every request.

</details>

<details>
<summary>Read the current tenant ID</summary>

`get_tenant_id` returns the tenant set right now, or `null` if none is set:

```php
$tenantId = \aikido\get_tenant_id();
```

</details>

<details>
<summary>Skip the check for specific queries</summary>

Some queries don't need a tenant filter, like an admin dashboard that counts across all tenants. Wrap them in `without_idor_protection`:

```php
// IDOR checks are skipped for the query in this callback
$result = \aikido\without_idor_protection(function () use ($pdo) {
    return $pdo->query("SELECT count(*) FROM agents WHERE status = 'running'");
});
```

</details>

## Troubleshooting

<details>
<summary>Missing tenant filter</summary>

```
Zen IDOR protection: query on table 'orders' is missing a filter on column 'tenant_id'
```

You have a query like `SELECT * FROM orders WHERE status = 'active'` with no `tenant_id` filter. Same check applies to `UPDATE` and `DELETE`.

</details>

<details>
<summary>Wrong tenant ID value</summary>

```
Zen IDOR protection: query on table 'orders' filters 'tenant_id' with value '456' but tenant ID is '123'
```

The query filters on `tenant_id`, but the value doesn't match the tenant set with `set_tenant_id`.

</details>

<details>
<summary>Missing tenant column in INSERT</summary>

```
Zen IDOR protection: INSERT on table 'orders' is missing column 'tenant_id'
```

An `INSERT` doesn't include the tenant column. Every INSERT needs it, with the right value.

</details>

<details>
<summary>Wrong tenant ID in INSERT</summary>

```
Zen IDOR protection: INSERT on table 'orders' sets 'tenant_id' to '456' but tenant ID is '123'
```

The INSERT has the tenant column, but the value doesn't match the tenant set with `set_tenant_id`.

</details>

<details>
<summary>Tenant ID could not be resolved</summary>

```
Zen IDOR protection: query on table 'orders' has a placeholder for 'tenant_id' that could not be resolved
```

The query filters on the tenant column, but Zen couldn't work out which value the placeholder stands for, so it can't tell whether the filter is correct. This happens when the value was never bound, when it's `null`, or when it isn't a plain string or number (an array or an object, for example). `INSERT` has its own version of this message.

</details>

<details>
<summary>Missing tenant ID</summary>

```
Zen IDOR protection: query on table 'orders' requires a tenant ID, but set_tenant_id() was not called.
```

Call `set_tenant_id` before running queries that need a tenant filter. Statements that don't touch a tenant-scoped table — migrations and other DDL, `BEGIN`/`COMMIT`, `SELECT 1` health checks, and queries that only read excluded tables — run fine before a tenant is set. That's what lets you look up the current user in a shared `users` table to work out which tenant to set. A query joining several tables lists all of them, e.g. `query on tables 'tickets, comments' requires a tenant ID`.

</details>

## Limitations

<details>
<summary><code>mysqli</code> prepared statements with <code>bind_param</code></summary>

`mysqli_stmt::bind_param` binds values natively in the driver, outside of the query string Zen sees, so it can't resolve them yet. Wrap these in `without_idor_protection()`, or switch the query to `mysqli_query`/`mysqli_real_query` with the values written directly into the query string.

</details>

## Statements that always pass

Zen only checks statements that read or change rows (`SELECT`, `INSERT`, `UPDATE`, `DELETE`). It recognizes these too, and they never trigger an IDOR error:

- DDL — `CREATE TABLE`, `ALTER TABLE`, `DROP TABLE`, ...
- Session commands — `SET`, `SHOW`, ...
- Transactions — `BEGIN`, `COMMIT`, `ROLLBACK`, ...
- Queries that touch no tenant-scoped table — `SELECT 1`, and any query reading only excluded tables

None of these need a tenant, so they also run before `set_tenant_id` has been called. That's what makes the usual startup order work: enable protection, run migrations, look the current user up in an excluded `users` table, then set the tenant you found.
