package cloud

import (
	"io"
	. "main/aikido_types"
	"main/log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestSendStartEventLogsFailureBeforeTraffic(t *testing.T) {
	cloudServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer cloudServer.Close()

	originalStdout := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to capture stdout: %v", err)
	}
	os.Stdout = writer
	defer func() {
		os.Stdout = originalStdout
		_ = writer.Close()
		_ = reader.Close()
	}()

	server := &ServerData{
		Logger: log.CreateLogger("test", "WARN", false),
		AikidoConfig: AikidoConfigData{
			Token:    "AIK_TEST",
			Endpoint: cloudServer.URL,
		},
	}

	SendStartEvent(server)
	closeErr := writer.Close()
	os.Stdout = originalStdout
	if closeErr != nil {
		t.Fatalf("failed to close captured stdout: %v", closeErr)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read captured stdout: %v", err)
	}
	if !strings.Contains(string(output), "Error in sending start event: received non-OK response: 503 Service Unavailable") {
		t.Fatalf("expected start event failure to be logged before traffic, got %q", output)
	}
}
