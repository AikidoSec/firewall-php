#pragma once

void LoadEnvironment();

void LoadSystemEnvironment();

bool LoadLaravelEnvFile();

bool GetBoolFromString(const std::string& env, bool default_value);

bool GetEnvBool(const std::string& env_key, bool default_value);

std::string GetEnvString(const std::string& env_key, const std::string default_value);
