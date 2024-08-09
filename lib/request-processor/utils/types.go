package utils

import "encoding/json"

type Kind string

const (
	Nosql_injection Kind = "nosql_injection"
	Sql_injection   Kind = "sql_injection"
	Shell_injection Kind = "shell_injection"
	Path_traversal  Kind = "path_traversal"
	Ssrf            Kind = "ssrf"
)

type InterceptorResult struct {
	Operation     string
	Kind          Kind
	Source        string
	PathToPayload string
	Metadata      map[string]string
	Payload       string
}

func (i InterceptorResult) ToString() string {
	json, _ := json.Marshal(i)
	return string(json)
}
