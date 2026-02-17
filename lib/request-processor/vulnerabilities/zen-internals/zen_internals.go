package zen_internals

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdlib.h>

typedef int (*detect_sql_injection_func)(const char*, size_t, const char*, size_t, int);
typedef int (*detect_shell_injection_func)(const char*, const char*);
typedef char* (*idor_analyze_sql_func)(const char*, size_t, int);
typedef void (*free_string_func)(char*);

int call_detect_shell_injection(detect_shell_injection_func func, const char* command, const char* user_input) {
    return func(command, user_input);
}

int call_detect_sql_injection(detect_sql_injection_func func,
                              const char* query, size_t query_len,
                              const char* input, size_t input_len,
                              int sql_dialect) {
    return func(query, query_len, input, input_len, sql_dialect);
}

char* call_idor_analyze_sql(idor_analyze_sql_func func,
                            const char* query, size_t query_len,
                            int dialect) {
    return func(query, query_len, dialect);
}

void call_free_string(free_string_func func, char* ptr) {
    func(ptr);
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
	idorAnalyzeSql     C.idor_analyze_sql_func
	freeString         C.free_string_func
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

	idorAnalyzeSqlFnName := C.CString("idor_analyze_sql_ffi")
	defer C.free(unsafe.Pointer(idorAnalyzeSqlFnName))

	vIdorAnalyzeSql := C.dlsym(handle, idorAnalyzeSqlFnName)
	if vIdorAnalyzeSql == nil {
		log.Error("Failed to load idor_analyze_sql_ffi function from zen-internals library!")
		return false
	}

	idorAnalyzeSql = (C.idor_analyze_sql_func)(vIdorAnalyzeSql)

	freeStringFnName := C.CString("free_string")
	defer C.free(unsafe.Pointer(freeStringFnName))

	vFreeString := C.dlsym(handle, freeStringFnName)
	if vFreeString == nil {
		log.Error("Failed to load free_string function from zen-internals library!")
		return false
	}

	freeString = (C.free_string_func)(vFreeString)

	log.Debugf("Loaded zen-internals library!")
	return true
}

func Uninit() {
	detectSqlInjection = nil
	idorAnalyzeSql = nil
	freeString = nil

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

	queryLen := C.size_t(len(query))
	userInputLen := C.size_t(len(user_input))

	result := int(C.call_detect_sql_injection(detectSqlInjection,
		cQuery, queryLen,
		cUserInput, userInputLen,
		C.int(dialect)))

	log.Debugf("DetectSqlInjection(\"%s\", \"%s\", %d) -> %d", query, user_input, dialect, result)
	return result
}

func IdorAnalyzeSql(query string, dialect int) string {
	if idorAnalyzeSql == nil || freeString == nil {
		return ""
	}

	cQuery := C.CString(query)
	defer C.free(unsafe.Pointer(cQuery))
	queryLen := C.size_t(len(query))

	resultPtr := C.call_idor_analyze_sql(idorAnalyzeSql, cQuery, queryLen, C.int(dialect))
	if resultPtr == nil {
		return ""
	}

	result := C.GoString(resultPtr)
	C.call_free_string(freeString, resultPtr)

	return result
}
