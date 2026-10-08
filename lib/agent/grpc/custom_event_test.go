package grpc

import (
	"context"
	. "main/aikido_types"
	"main/constants"
	"main/globals"
	"main/ipc/protos"
	"main/log"
	"testing"
	"time"
)

func TestOnCustomEventDoesNotWaitForCloudConfig(t *testing.T) {
	key := ServerKey{Token: "AIK_CUSTOM_EVENT_CONFIG_TEST", ServerPID: 493}
	server := &ServerData{Logger: log.CreateLogger("test", "ERROR", false)}
	globals.ServersMutex.Lock()
	globals.Servers[key] = server
	globals.ServersMutex.Unlock()
	defer func() {
		globals.ServersMutex.Lock()
		delete(globals.Servers, key)
		globals.ServersMutex.Unlock()
	}()

	server.CloudConfigMutex.Lock()
	defer server.CloudConfigMutex.Unlock()

	scheduled := make(chan error, 1)
	go func() {
		handler := &GrpcServer{}
		for i := 0; i <= constants.MaxConcurrentCustomEventRequests; i++ {
			_, err := handler.OnCustomEvent(context.Background(), &protos.CustomEvent{
				Token: key.Token, ServerPid: key.ServerPID, Name: "user.login_failed",
			})
			if err != nil {
				scheduled <- err
				return
			}
		}
		scheduled <- nil
	}()

	select {
	case err := <-scheduled:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("custom event scheduling must not wait for cloud configuration")
	}
}
