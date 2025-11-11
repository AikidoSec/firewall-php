package api_discovery

func onlyContainsPrimitiveTypes(types []string) bool {
	for _, item := range types {
		if !isPrimitiveType(item) {
			return false
		}
	}
	return true
}

func isPrimitiveType(typeStr string) bool {
	return typeStr != "object" && typeStr != "array"
}
