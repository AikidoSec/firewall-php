package aikido_types

import "main/ipc/protos"

type AITokens struct {
	Input  int64 `json:"input"`
	Output int64 `json:"output"`
	Total  int64 `json:"total"`
}

type AIStats struct {
	Provider string   `json:"provider"`
	Model    string   `json:"model"`
	Calls    int64    `json:"calls"`
	Tokens   AITokens `json:"tokens"`
}

type AIKey struct{ Provider, Model string }

func (server *ServerData) StoreAIUsage(call *protos.AIUsage) {
	if call == nil || call.Provider == "" || call.Model == "" {
		return
	}
	server.AIMutex.Lock()
	defer server.AIMutex.Unlock()
	if server.AI == nil {
		server.AI = make(map[AIKey]AIStats)
	}
	key := AIKey{Provider: call.Provider, Model: call.Model}
	stats, found := server.AI[key]
	if !found && len(server.AI) >= MaxNumberOfStoredAIStats {
		return
	}
	stats.Provider, stats.Model = call.Provider, call.Model
	stats.Calls++
	stats.Tokens.Input += call.Input
	stats.Tokens.Output += call.Output
	stats.Tokens.Total += call.Input + call.Output
	server.AI[key] = stats
}
