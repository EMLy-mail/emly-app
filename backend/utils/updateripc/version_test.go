package updateripc

import "testing"

func TestCheckPeerVersionV1(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"empty is rejected", "", true},
		{"below min is rejected", "1.1.9", true},
		{"unparseable is rejected", "not-a-version", true},
		{"exactly min is accepted", MinCompatibleUpdaterVersionV1, false},
		{"above max is accepted (informational only)", "9.9.9", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPeerVersionV1(tc.version)
			if (err != nil) != tc.wantErr {
				t.Errorf("checkPeerVersionV1(%q) error = %v, wantErr %v", tc.version, err, tc.wantErr)
			}
		})
	}
}

func TestCheckPeerVersionV2(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"empty is rejected", "", true},
		{"below min is rejected", "1.2.9", true},
		{"unparseable is rejected", "not-a-version", true},
		{"exactly min is accepted", MinCompatibleUpdaterVersionV2, false},
		{"above max is accepted (informational only)", "9.9.9", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPeerVersionV2(tc.version)
			if (err != nil) != tc.wantErr {
				t.Errorf("checkPeerVersionV2(%q) error = %v, wantErr %v", tc.version, err, tc.wantErr)
			}
		})
	}
}

func TestLess(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"1.1.0", "1.2.0b", true},
		{"1.2.0b", "1.3.0", true},
		{"1.8.0", "1.8.0", false},
		{"2.0.0", "1.9.9", false},
	}
	for _, tc := range tests {
		got, err := less(tc.a, tc.b)
		if err != nil {
			t.Fatalf("less(%q, %q) error: %v", tc.a, tc.b, err)
		}
		if got != tc.want {
			t.Errorf("less(%q, %q) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}
