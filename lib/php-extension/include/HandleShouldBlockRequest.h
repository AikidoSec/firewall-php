#pragma once

ZEND_BEGIN_ARG_INFO(arginfo_aikido_should_block_request, 0)
// No arguments
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_INFO(arginfo_aikido_auto_block_request, 0)
// No arguments
ZEND_END_ARG_INFO()

ZEND_FUNCTION(should_block_request);
ZEND_FUNCTION(auto_block_request);

void RegisterAikidoBlockRequestStatusClass();

bool get_blocking_status();
