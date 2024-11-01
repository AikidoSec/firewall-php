package sql_injection

// SQL keywords
var SQL_KEYWORDS = []string{
	"INSERT",
	"SELECT",
	"CREATE",
	"DROP",
	"DATABASE",
	"UPDATE",
	"DELETE",
	"ALTER",
	"GRANT",
	"SAVEPOINT",
	"COMMIT",
	"ROLLBACK",
	"TRUNCATE",
	"OR",
	"AND",
	"UNION",
	"AS",
	"WHERE",
	"DISTINCT",
	"FROM",
	"INTO",
	"TOP",
	"BETWEEN",
	"LIKE",
	"IN",
	"NULL",
	"NOT",
	"TABLE",
	"INDEX",
	"VIEW",
	"COUNT",
	"SUM",
	"AVG",
	"MIN",
	"MAX",
	"GROUP",
	"BY",
	"HAVING",
	"DESC",
	"ASC",
	"OFFSET",
	"FETCH",
	"LEFT",
	"RIGHT",
	"INNER",
	"OUTER",
	"JOIN",
	"EXISTS",
	"REVOKE",
	"ALL",
	"LIMIT",
	"ORDER",
	"ADD",
	"CONSTRAINT",
	"COLUMN",
	"ANY",
	"BACKUP",
	"CASE",
	"CHECK",
	"REPLACE",
	"DEFAULT",
	"EXEC",
	"FOREIGN",
	"KEY",
	"FULL",
	"PROCEDURE",
	"ROWNUM",
	"SET",
	"SESSION",
	"GLOBAL",
	"UNIQUE",
	"VALUES",
	"COLLATE",
	"IS",
}

// This is a list of common SQL keywords that are not dangerous by themselves
// They will appear in almost any SQL query
// e.g. SELECT * FROM table WHERE column = 'value' LIMIT 1
// If a query parameter is ?LIMIT=1 it would be blocked
// If the body contains "LIMIT" or "SELECT" it would be blocked
var COMMON_SQL_KEYWORDS = []string{
	"SELECT",
	"INSERT",
	"FROM",
	"WHERE",
	"DELETE",
	"GROUP",
	"BY",
	"ORDER",
	"LIMIT",
	"OFFSET",
	"HAVING",
	"COUNT",
	"SUM",
	"AVG",
	"MIN",
	"MAX",
	"DISTINCT",
	"AS",
	"AND",
	"OR",
	"NOT",
	"IN",
	"LIKE",
	"BETWEEN",
	"IS",
	"NULL",
	"ALL",
	"ANY",
	"EXISTS",
	"UNIQUE",
	"UPDATE",
	"INTO",
}

var SQL_OPERATORS = []string{
	"=",
	"!",
	";",
	"+",
	"-",
	"*",
	"/",
	"%",
	"&",
	"|",
	"^",
	">",
	"<",
	"#",
	"::",
}

var SQL_STRING_CHARS = []string{"\"", "'", "`"}

var SQL_DANGEROUS_IN_STRING = []string{
	"\"", // Double quote
	"'",  // Single quote
	"`",  // Backtick
	"\\", // Escape character
	"/*", // Start of comment
	"*/", // End of comment
	"--", // Start of comment
	"#",  // Start of comment

}

var SQL_ESCAPE_SEQUENCES = []string{"\\n", "\\r", "\\t"}
