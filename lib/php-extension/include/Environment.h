#pragma once

void LoadEnvironment();

void LoadSystemEnvironment();

bool LoadLaravelEnvFile();

// This should be used only after MINIT
bool GetEnvBool(const std::string& env_key, bool default_value);
