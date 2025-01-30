#pragma once

GoString GoCreateString(const std::string&);

GoSlice GoCreateSlice(const std::vector<int64_t>& v);

char* GoContextCallback(int callbackId);
