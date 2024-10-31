package zen_internals

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

typedef int (*detect_sql_injection_func)(const char*, const char*, int);
typedef int (*detect_shell_injection_func)(const char*, const char*);

int call_detect_shell_injection(detect_shell_injection_func func, const char* command, const char* user_input) {
    return func(command, user_input);
}

int call_detect_sql_injection(detect_sql_injection_func func, const char* query, const char* input, int sql_dialect) {
    return func(query, input, sql_dialect);
}
*/
import "C"
import (
	"errors"
	"fmt"
	"main/globals"
	"unsafe"
)

var (
	handle               unsafe.Pointer
	detectSqlInjection   C.detect_sql_injection_func
	detectShellInjection C.detect_shell_injection_func
)

func InitZenInternals() error {

	zenInternalsLibPath := "/opt/aikido-" + globals.Version + "/libzen_internals_x86_64-unknown-linux-gnu.so"
	handle := C.dlopen(C.CString(zenInternalsLibPath), C.RTLD_LAZY)
	fmt.Println(handle)
	if handle == nil {
		return errors.New("failed to load shared library")
	}

	// Retrieve the detect_sql_injection function pointer
	vDetectSqlInjection := C.dlsym(handle, C.CString("detect_sql_injection"))
	if vDetectSqlInjection == nil {
		return errors.New("failed to find detect_sql_injection function")
	}
	vDetectShellInjection := C.dlsym(handle, C.CString("detect_shell_injection"))
	if vDetectShellInjection == nil {
		return errors.New("failed to find detect_shell_injection function")
	}

	detectSqlInjection = (C.detect_sql_injection_func)(vDetectSqlInjection)
	detectShellInjection = (C.detect_shell_injection_func)(vDetectShellInjection)
	return nil
}

func CloseZenInternals() {
	if handle != nil {
		C.dlclose(handle)
		handle = nil
	}
}

// DetectSQLInjection performs SQL injection detection using the loaded library
func DetectSQLInjection(query string, user_input string, dialect int) (int, error) {
	if detectSqlInjection == nil {
		return 0, errors.New("detect_sql_injection function not initialized")
	}

	// Convert strings to C strings
	cQuery := C.CString(query)
	cUserInput := C.CString(user_input)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cUserInput))

	// Call the detect_sql_injection function
	result := int(C.call_detect_sql_injection(detectSqlInjection, cQuery, cUserInput, C.int(dialect)))
	return result, nil
}

// DetectShellInjection performs shell injection detection using the loaded library
func DetectShellInjection(command string, user_input string) (int, error) {
	if detectShellInjection == nil {
		return 0, errors.New("detect_shell_injection function not initialized")
	}

	// Convert strings to C strings
	cCommand := C.CString(command)
	cUserInput := C.CString(user_input)
	defer C.free(unsafe.Pointer(cCommand))
	defer C.free(unsafe.Pointer(cUserInput))

	// Call the detect_shell_injection function
	result := int(C.call_detect_shell_injection(detectShellInjection, cCommand, cUserInput))
	return result, nil
}
