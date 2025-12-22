---
favorite: true
---

# Requirements

### PHP versions
Zen for PHP supports the following PHP versions: 7.2, 7.3, 7.4, 8.0, 8.1, 8.2, 8.3, 8.4, 8.5.

### Web frameworks

Zen for PHP is Web-framework agnostic, meaning that it will work on any PHP Web framework that you want to use.

Zen for PHP can do this because the monitored functions are hooked at the PHP-core level.

### Database drivers
* ✅ [`PDO`](https://www.php.net/manual/en/book.pdo.php)
    * ✅ [`MySQL`](https://www.php.net/manual/en/ref.pdo-mysql.php)
    * ✅ [`Oracle`](https://www.php.net/manual/en/ref.pdo-oci.php)
    * ✅ [`PostgreSQL`](https://www.php.net/manual/en/ref.pdo-pgsql.php)
    * ✅ [`ODBC and DB2`](https://www.php.net/manual/en/ref.pdo-odbc.php)
    * ✅ [`Firebird`](https://www.php.net/manual/en/ref.pdo-firebird.php)
    * ✅ [`Microsoft SQL Server`](https://www.php.net/manual/en/ref.pdo-dblib.php)
    * ✅ [`SQLite`](https://www.php.net/manual/en/ref.pdo-sqlite.php)
* ✅ [`MySQLi`](https://www.php.net/manual/en/book.mysqli.php)
* 🚧 [`Oracle OCI8`](https://www.php.net/manual/en/book.oci8.php)
* 🚧 [`PostgreSQL`](https://www.php.net/manual/en/book.pgsql.php)
* 🚧 [`SQLite3`](https://www.php.net/manual/en/book.sqlite3.php)

### Outgoing requests libraries
* ✅ [`cURL`](https://www.php.net/manual/en/book.curl.php)
  * Including wrappers, when configured to use cURL as adapter:
    * ✅ [`GuzzleHttp`](https://docs.guzzlephp.org/en/stable/)
    * ✅ [`HTTP_Request2`](https://pear.php.net/package/http_request2)
    * ✅ [`Symfony\HTTPClient`](https://symfony.com/doc/current/http_client.html)
* ✅ [`file_get_contents`](https://www.php.net/manual/en/function.file-get-contents.php)
