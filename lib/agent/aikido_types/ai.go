package aikido_types

import "main/ipc/protos"

type AITokens struct {
	Input      int64 `json:"input"`
	Output     int64 `json:"output"`
	Total      int64 `json:"total"`
	CacheRead  int64 `json:"cacheRead"`
	CacheWrite int64 `json:"cacheWrite"`
}

type AIStats struct {
	Provider string   `json:"provider"`
	Model    string   `json:"model"`
	Calls    int64    `json:"calls"`
	Tokens   AITokens `json:"tokens"`
}

type AIKey struct{ Provider, Model string }

func (server *ServerData) StoreAIUsage(calls []*protos.AIUsage) {
	server.AIMutex.Lock()
	defer server.AIMutex.Unlock()
	if server.AI == nil {
		server.AI = make(map[AIKey]AIStats)
	}
	for _, call := range calls {
		if call == nil || call.Provider == "" || call.Model == "" {
			continue
		}
		key := AIKey{Provider: call.Provider, Model: call.Model}
		stats, found := server.AI[key]
		if !found && len(server.AI) >= 1000 {
			continue
		}
		stats.Provider, stats.Model = call.Provider, call.Model
		stats.Calls++
		stats.Tokens.Input += call.Input
		stats.Tokens.Output += call.Output
		stats.Tokens.Total += call.Input + call.Output
		stats.Tokens.CacheRead += call.CacheRead
		stats.Tokens.CacheWrite += call.CacheWrite
		server.AI[key] = stats
	}
}
