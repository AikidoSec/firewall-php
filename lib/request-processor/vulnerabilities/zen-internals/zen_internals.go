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
	"fmt"
	"main/globals"
	"main/log"
	"main/utils"
	"unsafe"
)

var (
	handle             unsafe.Pointer
	detectSqlInjection C.detect_sql_injection_func
)

func Init() bool {
	zenInternalsLibPath := C.CString(fmt.Sprintf("/opt/aikido-%s/libzen_internals_%s-unknown-linux-gnu.so", globals.Version, utils.GetArch()))
	defer C.free(unsafe.Pointer(zenInternalsLibPath))

	handle := C.dlopen(zenInternalsLibPath, C.RTLD_LAZY)
	if handle == nil {
		log.Errorf("Failed to load zen-internals library from '%s' with error %s!", C.GoString(zenInternalsLibPath), C.GoString(C.dlerror()))
		return false
	}

	detectSqlInjectionFnName := C.CString("detect_sql_injection")
	defer C.free(unsafe.Pointer(detectSqlInjectionFnName))

	vDetectSqlInjection := C.dlsym(handle, detectSqlInjectionFnName)
	if vDetectSqlInjection == nil {
		log.Error("Failed to load detect_sql_injection function from zen-internals library!")
		return false
	}

	detectSqlInjection = (C.detect_sql_injection_func)(vDetectSqlInjection)
	log.Debugf("Loaded zen-internals library!")
	return true
}

func Uninit() {
	detectSqlInjection = nil

	if handle != nil {
		C.dlclose(handle)
		handle = nil
	}
}

// DetectSQLInjection performs SQL injection detection using the loaded library
func DetectSQLInjection(query string, user_input string, dialect int) int {
	if detectSqlInjection == nil {
		return 0
	}

	// Convert strings to C strings
	cQuery := C.CString(query)
	cUserInput := C.CString(user_input)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cUserInput))

	// Call the detect_sql_injection function
	result := int(C.call_detect_sql_injection(detectSqlInjection, cQuery, cUserInput, C.int(dialect)))
	log.Debugf("DetectSqlInjection(%s, %s, %d) -> %d", query, user_input, dialect, result)
	return result
}
