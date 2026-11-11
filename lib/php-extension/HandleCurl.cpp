#include "Includes.h"
#include "HandleCurl.h"

#if PHP_VERSION_ID >= 80000
#include "Zend/zend_observer.h"

static bool AIEnabled() {
    return !AIKIDO_GLOBAL(disable) && AIKIDO_GLOBAL(ai_network_capture);
}

static bool EndsWith(const std::string& value, const std::string& suffix) {
    return value.size() >= suffix.size() && value.compare(value.size() - suffix.size(), suffix.size(), suffix) == 0;
}

static zval* CurlArgument(zend_execute_data* data, uint32_t index = 1) {
    if (ZEND_CALL_NUM_ARGS(data) < index) {
        return nullptr;
    }
    zval* handle = ZEND_CALL_ARG(data, index);
    if (Z_TYPE_P(handle) != IS_OBJECT || !zend_string_equals_literal(Z_OBJCE_P(handle)->name, "CurlHandle")) {
        return nullptr;
    }
    return handle;
}

static void StartTransfer(zval* handle) {
    if (!handle) {
        return;
    }
    auto& transfers = AIKIDO_GLOBAL(curlCapture).transfers;
    transfers.erase(Z_OBJ_HANDLE_P(handle));
    if (AIPath(CallPhpFunctionCurlGetInfo(handle, CURLINFO_EFFECTIVE_URL))) {
        CurlTransfer transfer;
        auto& callbacks = AIKIDO_GLOBAL(curlCapture).callbacks;
        auto callback = callbacks.find(Z_OBJ_HANDLE_P(handle));
        if (callback != callbacks.end()) {
            transfer.callback = callback->second;
        }
        transfers.emplace(Z_OBJ_HANDLE_P(handle), std::move(transfer));
    }
}

static void FinishTransfer(zval* handle, bool success, zval* returned = nullptr) {
    if (!handle || Z_TYPE_P(handle) != IS_OBJECT) {
        return;
    }
    auto& capture = AIKIDO_GLOBAL(curlCapture);
    auto it = capture.transfers.find(Z_OBJ_HANDLE_P(handle));
    // A redirect may lead to an AI URL even if we did not track the original URL.
    CurlTransfer transfer;
    if (it != capture.transfers.end()) {
        transfer = std::move(it->second);
        capture.transfers.erase(it);
    }
    if (!success || EG(exception) || transfer.discarded) {
        return;
    }
    if (returned && Z_TYPE_P(returned) == IS_STRING) {
        if (Z_STRLEN_P(returned) > MAX_AI_BODY_BYTES) {
            return;
        }
        transfer.body.assign(Z_STRVAL_P(returned), Z_STRLEN_P(returned));
    }
    if (transfer.body.empty() || transfer.body.size() > MAX_AI_BODY_BYTES) {
        return;
    }
#if LIBCURL_VERSION_NUM >= 0x074800
    if (CallPhpFunctionCurlGetInfo(handle, CURLINFO_EFFECTIVE_METHOD) != "POST") {
        return;
    }
#else
    return;
#endif
    auto status = CallPhpFunctionCurlGetInfo(handle, CURLINFO_RESPONSE_CODE);
    if (status.size() != 3 || status[0] != '2') {
        return;
    }
    auto contentType = ToLowercase(CallPhpFunctionCurlGetInfo(handle, CURLINFO_CONTENT_TYPE));
    contentType = contentType.substr(0, contentType.find(';'));
    if (contentType != "application/json" && !EndsWith(contentType, "+json")) {
        return;
    }
    auto url = CallPhpFunctionCurlGetInfo(handle, CURLINFO_EFFECTIVE_URL);
    if (!AIPath(url)) {
        return;
    }
    CaptureAIResponse(url, transfer.body);
}

static void ObserveCurlCallback(zend_execute_data* data) {
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        auto* func = data->func;
        auto* handle = CurlArgument(data);
        if (!handle || ZEND_CALL_NUM_ARGS(data) != 2) {
            return;
        }
        // The user's code can call the CURLOPT_WRITEFUNCTION callback directly. Only collect bytes when cURL calls it.
        auto* parent = data->prev_execute_data;
        if (!parent || !parent->func->common.function_name) {
            return;
        }
        auto* callerName = parent->func->common.function_name;
        bool calledByCurl = zend_string_equals_literal(callerName, "curl_exec") ||
                            zend_string_equals_literal(callerName, "curl_multi_exec");
        if (!calledByCurl) {
            return;
        }
        auto& transfers = AIKIDO_GLOBAL(curlCapture).transfers;
        auto it = transfers.find(Z_OBJ_HANDLE_P(handle));
        if (it == transfers.end() || it->second.discarded || it->second.callback != func) {
            return;
        }
        auto* bytes = ZEND_CALL_ARG(data, 2);
        if (Z_TYPE_P(bytes) != IS_STRING) {
            return;
        }
        // One chunk may contain only part of the JSON. Wait until the request finishes.
        auto& transfer = it->second;
        if (Z_STRLEN_P(bytes) > MAX_AI_BODY_BYTES - transfer.body.size()) {
            transfer.body.clear();
            transfer.discarded = true;
        } else {
            transfer.body.append(Z_STRVAL_P(bytes), Z_STRLEN_P(bytes));
        }
    } catch (...) {}
}

static void CopyCallback(zend_execute_data* data, zval* returned) {
    auto* source = CurlArgument(data);
    if (!source || Z_TYPE_P(returned) != IS_OBJECT) {
        return;
    }
    auto& callbacks = AIKIDO_GLOBAL(curlCapture).callbacks;
    auto callback = callbacks.find(Z_OBJ_HANDLE_P(source));
    if (callback != callbacks.end() && callbacks.size() < 128) {
        callbacks[Z_OBJ_HANDLE_P(returned)] = callback->second;
    }
}

// curl_multi_info_read() removes a message from the queue.
// Use the message already read by the user's code instead of reading another one.
static void FinishMultiTransfer(zval* returned) {
    if (Z_TYPE_P(returned) != IS_ARRAY) {
        return;
    }
    auto* msg = zend_hash_str_find(Z_ARRVAL_P(returned), ZEND_STRL("msg"));
    auto* result = zend_hash_str_find(Z_ARRVAL_P(returned), ZEND_STRL("result"));
    auto* handle = zend_hash_str_find(Z_ARRVAL_P(returned), ZEND_STRL("handle"));
    if (msg && result && handle && Z_TYPE_P(msg) == IS_LONG && Z_LVAL_P(msg) == CURLMSG_DONE && Z_TYPE_P(result) == IS_LONG) {
        zval body;
        ZVAL_UNDEF(&body);
        // getcontent reads the body without clearing it. The user's code can still read it later.
        if (Z_LVAL_P(result) == CURLE_OK) {
            CallPhpFunction("curl_multi_getcontent", 1, handle, &body);
        }
        try {
            FinishTransfer(handle, Z_LVAL_P(result) == CURLE_OK, &body);
        } catch (...) {
            zval_ptr_dtor(&body);
            throw;
        }
        zval_ptr_dtor(&body);
    }
}

static void TrackCallback(zend_execute_data* data, zval* returned) {
    if (Z_TYPE_P(returned) != IS_TRUE) {
        return;
    }
    auto* handle = CurlArgument(data);
    if (!handle) {
        return;
    }
    zval* value = nullptr;
    if (zend_string_equals_literal(data->func->common.function_name, "curl_setopt") && ZEND_CALL_NUM_ARGS(data) == 3) {
        auto* option = ZEND_CALL_ARG(data, 2);
        if (Z_TYPE_P(option) == IS_LONG && Z_LVAL_P(option) == CURLOPT_WRITEFUNCTION) {
            value = ZEND_CALL_ARG(data, 3);
        }
    } else if (ZEND_CALL_NUM_ARGS(data) == 2 && Z_TYPE_P(ZEND_CALL_ARG(data, 2)) == IS_ARRAY) {
        value = zend_hash_index_find(Z_ARRVAL_P(ZEND_CALL_ARG(data, 2)), CURLOPT_WRITEFUNCTION);
    }
    if (!value) {
        return;
    }
    auto& callbacks = AIKIDO_GLOBAL(curlCapture).callbacks;
    callbacks.erase(Z_OBJ_HANDLE_P(handle));
    const zend_function* callback = nullptr;
    if (Z_TYPE_P(value) != IS_NULL && callbacks.size() < 128) {
        zend_fcall_info fci;
        zend_fcall_info_cache fcc = {};
        if (zend_fcall_info_init(value, 0, &fci, &fcc, nullptr, nullptr) == SUCCESS) {
            bool usesTrampoline = (fcc.function_handler->common.fn_flags & ZEND_ACC_CALL_VIA_TRAMPOLINE) != 0;
            // Skip magic-method callbacks: zend_release_fcall_info_cache() below deletes their handler.
            // Saving that pointer would leave us pointing to freed memory.
            if (!usesTrampoline) {
                callback = fcc.function_handler;
                callbacks[Z_OBJ_HANDLE_P(handle)] = callback;
            }
            zend_release_fcall_info_cache(&fcc);
        }
    }
    auto& transfers = AIKIDO_GLOBAL(curlCapture).transfers;
    auto transfer = transfers.find(Z_OBJ_HANDLE_P(handle));
    // After the first chunk, Symfony replaces the callback with curl_setopt($ch, \CURLOPT_WRITEFUNCTION, ...).
    // https://github.com/symfony/http-client/blob/5.4/Response/CurlResponse.php#L136-L153
    if (transfer != transfers.end()) {
        transfer->second.callback = callback;
    }
}

static zend_observer_fcall_handlers CurlObserver(zend_execute_data* data) {
    auto* func = data->func;
    if (func->type == ZEND_USER_FUNCTION) {
        return {ObserveCurlCallback, nullptr};
    }
    return {nullptr, nullptr};
}
#endif

void RegisterCurlObservers() {
    // PHP only tells observers about callbacks called from C starting with PHP 8.2.
#if PHP_VERSION_ID >= 80200
    zend_observer_fcall_register(CurlObserver);
#endif
}

void CurlCapture::Reset() {
    transfers.clear();
    callbacks.clear();
}

AIKIDO_HANDLER_FUNCTION(handle_pre_curl_ai_capture) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        StartTransfer(CurlArgument(execute_data));
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_ai_capture) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        FinishTransfer(CurlArgument(execute_data), Z_TYPE_P(return_value) != IS_FALSE, return_value);
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_pre_curl_multi_add_handle) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        StartTransfer(CurlArgument(execute_data, 2));
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_multi_add_handle) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        if (Z_TYPE_P(return_value) != IS_LONG || Z_LVAL_P(return_value) != CURLM_OK) {
            auto* handle = CurlArgument(execute_data, 2);
            if (handle) {
                AIKIDO_GLOBAL(curlCapture).transfers.erase(Z_OBJ_HANDLE_P(handle));
            }
        }
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_multi_info_read) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        FinishMultiTransfer(return_value);
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_setopt) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        TrackCallback(execute_data, return_value);
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_setopt_array) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        TrackCallback(execute_data, return_value);
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_copy_handle) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        CopyCallback(execute_data, return_value);
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_pre_curl_reset) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        auto* handle = CurlArgument(execute_data);
        if (handle) {
            AIKIDO_GLOBAL(curlCapture).transfers.erase(Z_OBJ_HANDLE_P(handle));
            AIKIDO_GLOBAL(curlCapture).callbacks.erase(Z_OBJ_HANDLE_P(handle));
        }
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_pre_curl_close) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        auto* handle = CurlArgument(execute_data);
        if (handle) {
            AIKIDO_GLOBAL(curlCapture).transfers.erase(Z_OBJ_HANDLE_P(handle));
            AIKIDO_GLOBAL(curlCapture).callbacks.erase(Z_OBJ_HANDLE_P(handle));
        }
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_post_curl_multi_remove_handle) {
#if PHP_VERSION_ID >= 80000
    if (!AIEnabled() || EG(exception)) {
        return;
    }
    try {
        auto* handle = CurlArgument(execute_data, 2);
        if (handle) {
            AIKIDO_GLOBAL(curlCapture).transfers.erase(Z_OBJ_HANDLE_P(handle));
        }
    } catch (...) {}
#endif
}

AIKIDO_HANDLER_FUNCTION(handle_curl_capture_noop) {}
