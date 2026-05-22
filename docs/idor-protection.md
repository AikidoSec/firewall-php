# IDOR Protection

IDOR stands for Insecure Direct Object Reference — it's when one account can access another account's data because a query doesn't properly filter by account.

If your SaaS has accounts (or organizations, workspaces, teams, ...) and uses a column like `tenant_id` to keep each account's data separate, IDOR protection ensures every SQL query filters on the correct tenant. Zen analyzes queries at runtime and throws an error if a query is missing that filter or uses the wrong tenant ID, catching mistakes like:

- A `SELECT` that forgets the tenant filter, letting one account read another's orders
- An `UPDATE` or `DELETE` without a tenant filter, letting one account modify another's data
- An `INSERT` that omits the tenant column, creating orphaned or misassigned rows

Zen catches these at runtime so they surface during development and testing, not in production. See [IDOR vulnerability explained](https://www.aikido.dev/blog/idor-vulnerability-explained) for more background.

> [!IMPORTANT]
> IDOR protection always throws an exception on violations regardless of block/detect mode. A missing filter is a developer bug, not an external attack.

## Setup

### 1. Enable IDOR protection at startup

```php
\aikido\enable_idor_protection("tenant_id", ["users"]);
```

- `tenant_column_name` — the column name that identifies the tenant in your database tables (e.g. `account_id`, `organization_id`, `team_id`).
- `excluded_tables` — tables that Zen should skip IDOR checks for, because rows aren't scoped to a single tenant (e.g. a shared `users` table that stores users across all tenants).

### 2. Set the tenant ID per request

Every request must have a tenant ID when IDOR protection is enabled. Call `set_tenant_id` early in your request handler (e.g. in middleware after authentication):

```php
\aikido\set_tenant_id($user->organizationId);
```

> [!IMPORTANT]
> If `set_tenant_id` is not called for a request, Zen will throw an exception when a SQL query is executed.

### 3. Bypass for specific queries (optional)

Some queries don't need tenant filtering (e.g. aggregations across all tenants for an admin dashboard). Use `without_idor_protection` to bypass the check for a specific callback:

```php
$result = \aikido\without_idor_protection(function() use ($pdo) {
    return $pdo->query("SELECT count(*) FROM agents WHERE status = 'running'");
});
```

## Troubleshooting

<details>
<summary>Missing tenant filter</summary>

```
Zen IDOR protection: query on table 'orders' is missing a filter on column 'tenant_id'
```

This means you have a query like `SELECT * FROM orders WHERE status = 'active'` that doesn't filter on `tenant_id`. The same check applies to `UPDATE` and `DELETE` queries.

</details>

<details>
<summary>Wrong tenant ID value</summary>

```
Zen IDOR protection: query on table 'orders' filters 'tenant_id' with value '456' but tenant ID is '123'
```

This means the query filters on `tenant_id`, but the value doesn't match the tenant ID set via `set_tenant_id`.

</details>

<details>
<summary>Missing tenant column in INSERT</summary>

```
Zen IDOR protection: INSERT on table 'orders' is missing column 'tenant_id'
```

This means an `INSERT` statement doesn't include the tenant column. Every INSERT must include the tenant column with the correct tenant ID value.

</details>

<details>
<summary>Wrong tenant ID in INSERT</summary>

```
Zen IDOR protection: INSERT on table 'orders' sets 'tenant_id' to '456' but tenant ID is '123'
```

This means the INSERT includes the tenant column, but the value doesn't match the tenant ID set via `set_tenant_id`.

</details>

<details>
<summary>Missing set_tenant_id call</summary>

```
Zen IDOR protection: setTenantId() was not called for this request. Every request must have a tenant ID when IDOR protection is enabled.
```

</details>

## Supported databases

- MySQL (via `mysqli` and PDO)
- PostgreSQL (via PDO)
- SQLite (via PDO)

Any ORM or query builder that uses PDO or mysqli under the hood is supported (e.g. Eloquent, Doctrine DBAL).

## Prepared statements

Zen supports placeholder resolution for prepared statements executed via `PDOStatement::execute()`:

```php
// Positional placeholders — values are resolved
$stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = ?");
$stmt->execute([$tenantId]);

// Named placeholders — values are resolved
$stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
$stmt->execute([':tid' => $tenantId]);
```

Parameters bound with `bindValue()` or `bindParam()` before calling `execute()` without arguments are also resolved:

```php
$stmt = $pdo->prepare("SELECT * FROM orders WHERE tenant_id = :tid");
$stmt->bindValue(':tid', $tenantId);
$stmt->execute();
```

## Statements that are always allowed

Zen only checks statements that read or modify row data (`SELECT`, `INSERT`, `UPDATE`, `DELETE`). The following statement types are also recognized and never trigger an IDOR error:

- DDL — `CREATE TABLE`, `ALTER TABLE`, `DROP TABLE`, ...
- Session commands — `SET`, `SHOW`, ...
- Transactions — `BEGIN`, `COMMIT`, `ROLLBACK`, ...
