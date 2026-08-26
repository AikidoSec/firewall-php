#pragma once

GoString GoCreateString(const std::string&);

GoSlice GoCreateSlice(const std::vector<int64_t>& v);

CallbackResult GoContextCallback(int callbackId);
