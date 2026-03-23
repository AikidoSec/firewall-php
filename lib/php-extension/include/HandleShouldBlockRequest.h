#pragma once

ZEND_BEGIN_ARG_INFO(arginfo_aikido_should_block_request, 0)
// No arguments
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_INFO(arginfo_aikido_auto_block_request, 0)
// No arguments
ZEND_END_ARG_INFO()

ZEND_BEGIN_ARG_INFO(arginfo_aikido_should_whitelist_request, 0)
// No arguments
ZEND_END_ARG_INFO()

// Function called by the users of Aikido, in order to check if a request should be blocked
// based on user information provided via set_user or if rate limiting exceeded.
ZEND_FUNCTION(should_block_request);

// Function call automatically injected by Aikido in the PHP code,
// in order to automatically block requests based on IP and User-Agent.
ZEND_FUNCTION(auto_block_request);

// Function called by the users of Aikido, in order to check if a request should be whitelisted
// based on IP.
ZEND_FUNCTION(should_whitelist_request);


void RegisterAikidoBlockRequestStatusClass();
void RegisterAikidoWhitelistRequestStatusClass();
