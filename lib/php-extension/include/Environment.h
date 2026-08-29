#pragma once

void LoadEnvironment();

void LoadSystemEnvironment();

bool LoadLaravelEnvFile();

bool GetEnvBoolWithAllGetters(const std::string& env_key, bool default_value);

std::string GetGuardEndpointWithAllGetters(const std::string& token);
std::string GetRealtimeEndpointWithAllGetters(const std::string& token);
