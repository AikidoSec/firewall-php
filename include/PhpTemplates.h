#pragma once

#define PHP_EXIT_ACTION_TEMPLATE "ob_clean();\n" \
                                 "header_remove();\n" \
                                 "http_response_code(%d);\n" \
                                 "header('Content-Type: text/plain');\n" \
                                 "echo '%s';\n" \
                                 "%s"
