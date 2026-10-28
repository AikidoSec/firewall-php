package aikido_types

import "testing"

func TestRequestShutdownParamsShouldReport(t *testing.T) {
	for _, tc := range []struct {
		name   string
		params RequestShutdownParams
		want   bool
	}{
		{"plain request", RequestShutdownParams{}, false},
		{"rate limited", RequestShutdownParams{RateLimited: true}, true},
		{"route discovery", RequestShutdownParams{ShouldDiscoverRoute: true}, true},
		{"web scanner", RequestShutdownParams{IsWebScanner: true}, true},
		{"php fatal error", RequestShutdownParams{PhpFatalErrorKind: "memory_exhausted"}, true},
		{"upload progress field", RequestShutdownParams{HasUploadProgressField: true}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.params.ShouldReport(); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
