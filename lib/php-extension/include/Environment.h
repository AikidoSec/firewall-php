#pragma once

void LoadEnvironment();

void LoadSystemEnvironment();

bool LoadDotEnvFile();

bool GetEnvBoolWithAllGetters(const std::string& env_key, bool default_value);

std::string GetGuardEndpointWithAllGetters(const std::string& token);
