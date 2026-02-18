package context

// #include "../../API.h"
// #include <pthread.h>
// static unsigned long get_thread_id() { return (unsigned long)pthread_self(); }
import "C"
import (
	"encoding/json"
	"fmt"
	. "main/aikido_types"
	"main/instance"
)

var TestContext map[string]string
var TestServer *ServerData // Test server for unit tests

func UnitTestsCallback(instance *instance.RequestProcessorInstance, context_id int) string {
	switch context_id {
	case C.CONTEXT_REMOTE_ADDRESS:
		return TestContext["remoteAddress"]
	case C.CONTEXT_HTTPS:
		return TestContext["https"]
	case C.CONTEXT_METHOD:
		return TestContext["method"]
	case C.CONTEXT_ROUTE:
		return TestContext["route"]
	case C.CONTEXT_URL:
		return TestContext["url"]
	case C.CONTEXT_QUERY:
		return TestContext["query"]
	case C.CONTEXT_STATUS_CODE:
		return TestContext["statusCode"]
	case C.CONTEXT_HEADERS:
		return TestContext["headers"]
	case C.CONTEXT_HEADER_X_FORWARDED_FOR:
		return TestContext["xForwardedFor"]
	case C.CONTEXT_HEADER_USER_AGENT:
		return TestContext["userAgent"]
	case C.CONTEXT_COOKIES:
		return TestContext["cookies"]
	case C.CONTEXT_BODY:
		return TestContext["body"]
	}
	return ""
}

func getThreadID() uint64 {
	return uint64(C.get_thread_id())
}

func LoadForUnitTests(context map[string]string) *instance.RequestProcessorInstance {
	tid := getThreadID()

	mockInst := instance.NewRequestProcessorInstance(tid)
	if TestServer != nil {
		mockInst.SetCurrentServer(TestServer)
		mockInst.SetCurrentToken(TestServer.AikidoConfig.Token)
	}

	ctx := &RequestContextData{
		instance: mockInst,
		Callback: UnitTestsCallback,
	}
	mockInst.SetRequestContext(ctx)
	mockInst.SetContextInstance(nil)
	mockInst.SetEventContext(&EventContextData{})

	TestContext = context
	return mockInst
}

func UnloadForUnitTests() {
	TestServer = nil
	TestContext = nil
}

func SetTestServer(instance *instance.RequestProcessorInstance, server *ServerData) {
	TestServer = server

	c := GetContext(instance)
	if c != nil && c.instance != nil && server != nil {
		c.instance.SetCurrentServer(server)
		c.instance.SetCurrentToken(server.AikidoConfig.Token)
	}
}

// GetTestServer returns the current test server, or nil if not set
func GetTestServer() *ServerData {
	return TestServer
}

func GetJsonString(m map[string]interface{}) string {
	jsonStr, err := json.Marshal(m)
	if err != nil {
		fmt.Println("Error converting map to JSON:", err)
		return ""
	}

	return string(jsonStr)
}
