#pragma once

#include "Includes.h"

/* Handles PHP functions that have a file path as first parameter (before) */
AIKIDO_HANDLER_FUNCTION(handle_pre_file_path_access);

/* Handles PHP functions that have a file path as first parameter (after) */
AIKIDO_HANDLER_FUNCTION(handle_post_file_path_access);

/* Handles PHP functions that have a file path as both first and second parameter (before) */
AIKIDO_HANDLER_FUNCTION(handle_pre_file_path_access_2);

/* Handles PHP functions that have a file path as both first and second parameter (after) */
AIKIDO_HANDLER_FUNCTION(handle_post_file_path_access_2);
