# Blocking invalid SQL queries

Zen can block SQL queries that it can't tokenize when they contain user input. This prevents attackers from bypassing SQL injection detection with malformed queries. For example, ClickHouse ignores invalid SQL after `;`, and SQLite runs queries before an unclosed `/*` comment.

This is NOT on by default, but it can be enabled by setting the `AIKIDO_BLOCK_INVALID_SQL` environment variable to `true`.
If enabled, in blocking mode, these queries are blocked. In detection-only mode, they are reported but still executed.

