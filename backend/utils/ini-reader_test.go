package utils

import "testing"

func TestIsValidGUIVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{"plain semver", "2.0.1", true},
		{"semver with prerelease", "2.0.1-beta.1", true},
		{"empty string is rejected", "", false},
		{"non-numeric is rejected", "latest", false},
		{"already v-prefixed is tolerated", "v2.0.1", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidGUIVersion(tc.version)
			if got != tc.want {
				t.Errorf("isValidGUIVersion(%q) = %v, want %v", tc.version, got, tc.want)
			}
		})
	}
}
