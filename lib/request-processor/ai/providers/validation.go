package providers

func validTokenCounts(counts ...int64) bool {
	for _, count := range counts {
		if count < 0 {
			return false
		}
	}
	return true
}
