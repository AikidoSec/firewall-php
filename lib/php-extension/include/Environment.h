#pragma once

void LoadEnvironment();

void LoadProcessEnvironment();

void LoadDotEnvFile();

bool GetEnvBoolWithAllGetters(const std::string& env_key, bool default_value);

std::string GetGuardEndpointWithAllGetters(const std::string& token);
