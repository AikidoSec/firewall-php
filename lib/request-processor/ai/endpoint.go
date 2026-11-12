package ai

import (
	"net/url"
	"strings"
)

type responseKind uint8

const (
	unknownEndpoint responseKind = iota
	openAIChat
	openAIResponses
	anthropic
	gemini
	bedrockConverse
	bedrockInvoke
)

type endpoint struct {
	kind   responseKind
	origin string
	model  string
}

func matchEndpoint(rawURL string) endpoint {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return endpoint{}
	}
	matched := classifyPath(u.EscapedPath())
	if matched.kind == unknownEndpoint {
		return matched
	}
	host := strings.ToLower(u.Host)
	if port := u.Port(); port == "0" || (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		host = strings.TrimSuffix(host, ":"+port)
	}
	matched.origin = u.Scheme + "://" + host
	return matched
}

func classifyPath(path string) endpoint {
	switch {
	case strings.HasSuffix(path, "/chat/completions"):
		return endpoint{kind: openAIChat}
	case strings.HasSuffix(path, "/responses"):
		return endpoint{kind: openAIResponses}
	case strings.HasSuffix(path, "/v1/messages"):
		return endpoint{kind: anthropic}
	}
	if model := modelInPath(path, "/publishers/anthropic/models/", ":rawPredict"); model != "" {
		return endpoint{kind: anthropic}
	}
	if model := modelInPath(path, "/models/", ":generateContent"); model != "" && !strings.Contains(model, ":") {
		return endpoint{kind: gemini, model: model}
	}
	if model := modelInPath(path, "/model/", "/converse"); model != "" {
		return endpoint{kind: bedrockConverse, model: model}
	}
	if model := modelInPath(path, "/model/", "/invoke"); model != "" {
		return endpoint{kind: bedrockInvoke, model: model}
	}
	return endpoint{}
}

func modelInPath(path, prefix, suffix string) string {
	path, ok := strings.CutSuffix(path, suffix)
	if !ok {
		return ""
	}
	start := strings.LastIndex(path, prefix)
	if start == -1 {
		return ""
	}
	model := path[start+len(prefix):]
	if strings.Contains(model, "/") {
		return ""
	}
	return model
}
