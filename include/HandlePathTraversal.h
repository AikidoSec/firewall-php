#pragma once

#include "Includes.h"

/* Handles PHP functions that have a file path as first parameter */
AIKIDO_HANDLER_FUNCTION(handle_file_path_access);

/* Handles PHP functions that have file paths as first and second parameter */
AIKIDO_HANDLER_FUNCTION(handle_file_path_access_2);