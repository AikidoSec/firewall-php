package zen_internals

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

typedef int (*detect_sql_injection_func)(const char*, size_t, const char*, size_t, int);
typedef int (*detect_shell_injection_func)(const char*, const char*);

int call_detect_shell_injection(detect_shell_injection_func func, const char* command, const char* user_input) {
    return func(command, user_input);
}

int call_detect_sql_injection(detect_sql_injection_func func,
                              const char* query, size_t query_len,
                              const char* input, size_t input_len,
                              int sql_dialect) {
    return func(query, query_len, input, input_len, sql_dialect);
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

type ZenInternalsLibrary struct {
	handle             unsafe.Pointer
	detectSqlInjection C.detect_sql_injection_func
}

var zenLib = &ZenInternalsLibrary{}

func Init() bool {
	zenInternalsLibPath := C.CString(fmt.Sprintf("/opt/aikido-%s/libzen_internals_%s-unknown-linux-gnu.so", globals.Version, utils.GetArch()))
	defer C.free(unsafe.Pointer(zenInternalsLibPath))

	handle := C.dlopen(zenInternalsLibPath, C.RTLD_LAZY)
	if handle == nil {
		log.Errorf(nil, "Failed to load zen-internals library from '%s' with error %s!", C.GoString(zenInternalsLibPath), C.GoString(C.dlerror()))
		return false
	}

	detectSqlInjectionFnName := C.CString("detect_sql_injection")
	defer C.free(unsafe.Pointer(detectSqlInjectionFnName))

	vDetectSqlInjection := C.dlsym(handle, detectSqlInjectionFnName)
	if vDetectSqlInjection == nil {
		log.Error(nil, "Failed to load detect_sql_injection function from zen-internals library!")
		C.dlclose(handle)
		return false
	}

	zenLib.handle = handle
	zenLib.detectSqlInjection = (C.detect_sql_injection_func)(vDetectSqlInjection)
	log.Debugf(nil, "Loaded zen-internals library!")
	return true
}

func Uninit() {
	zenLib.detectSqlInjection = nil

	if zenLib.handle != nil {
		C.dlclose(zenLib.handle)
		zenLib.handle = nil
	}
}

// DetectSQLInjection performs SQL injection detection using the loaded library
func DetectSQLInjection(query string, user_input string, dialect int) int {
	detectFn := zenLib.detectSqlInjection

	if detectFn == nil {
		return 0
	}

	// Convert strings to C strings
	cQuery := C.CString(query)
	cUserInput := C.CString(user_input)
	defer C.free(unsafe.Pointer(cQuery))
	defer C.free(unsafe.Pointer(cUserInput))

	queryLen := C.size_t(len(query))
	userInputLen := C.size_t(len(user_input))

	result := int(C.call_detect_sql_injection(detectFn,
		cQuery, queryLen,
		cUserInput, userInputLen,
		C.int(dialect)))

	log.Debugf(nil, "DetectSqlInjection(\"%s\", \"%s\", %d) -> %d", query, user_input, dialect, result)
	return result
}
