package config

import (
	"os"
	"testing"
)

func TestGetServerPID(t *testing.T) {
	currentPID := int32(os.Getpid())
	parentPID := int32(os.Getppid())

	tests := []struct {
		name         string
		platformName string
		expectedPID  int32
	}{
		{name: "PHP-FPM uses its master", platformName: "fpm-fcgi", expectedPID: parentPID},
		{name: "Apache uses its master", platformName: "apache2handler", expectedPID: parentPID},
		{name: "PHP built-in server uses itself", platformName: "cli-server", expectedPID: currentPID},
		{name: "FrankenPHP uses itself", platformName: "frankenphp", expectedPID: currentPID},
		{name: "unknown SAPI uses itself", platformName: "unknown", expectedPID: currentPID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualPID := getServerPID(test.platformName)
			if actualPID != test.expectedPID {
				t.Fatalf("getServerPID(%q) = %d, expected %d", test.platformName, actualPID, test.expectedPID)
			}
		})
	}
}
