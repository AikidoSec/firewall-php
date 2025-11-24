#pragma once

#include "Includes.h"

std::string ToLowercase(const std::string& str);

std::string ToUppercase(const std::string& str);

std::string GetRandomNumber();

std::string GetTime();

std::string GetDateTime();

const char* GetEventName(EVENT_ID event);

std::string NormalizeAndDumpJson(const json& jsonStr);

std::string ArrayToJson(zval* array);

std::string GetSqlDialectFromPdo(zval *pdo_object);

bool StartsWith(const std::string& str, const std::string& prefix, bool caseSensitive = true);

json CallPhpFunctionParseUrl(const std::string& url);

std::string AnonymizeToken(const std::string& str);

bool FileExists(const std::string& filePath);

bool RemoveFile(const std::string& filePath);

std::string GetStackTrace();

static inline zend_class_entry* GetFirewallDefaultExceptionCe();