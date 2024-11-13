package aikido_types

type HandlerFunction func() string

type Method struct {
	ClassName  string
	MethodName string
}

type QueryExecuted struct {
	Query     string
	Operation string
	Dialect   string
}

type FileAccessed struct {
	Filename  string
	Operation string
}

type ShellExecuted struct {
	Cmd       string
	Operation string
}
