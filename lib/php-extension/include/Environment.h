#pragma once

void LoadEnvironment();

bool LoadLaravelEnvFile();

bool GetBoolFromString(const std::string& env, bool default_value);
