#pragma once

#include "Includes.h"

std::string ToLowercase(const std::string& str);

std::string GetRandomNumber();

std::string GetTime();

std::string GetDateTime();

std::string GenerateSocketPath();

const char* GetEventName(EVENT_ID event);

std::string NormalizeAndDumpJson(const json& jsonStr);

std::string ArrayToJson(zval* array);

std::string GetSqlDialectFromPdo(zval *pdo_object);

bool StartsWith(const std::string& str, const std::string& prefix, bool caseSensitive = true);

json CallPhpFunctionParseUrl(const std::string& url);