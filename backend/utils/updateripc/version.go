package updateripc

import (
	"fmt"

	goversion "github.com/hashicorp/go-version"

	"emly/backend/utils"
)

// MinCompatibleUpdaterVersion / MaxCompatibleUpdaterVersion bound the
// EMLyUpdater releases known to speak protocolVersion — see the
// compatibility matrix atop proto/updateripc.proto. A response whose
// sender_version is below Min is rejected even though protocol_version
// matched, since not every required fix changes the wire format. Max is
// informational only (never enforced): a newer EMLyUpdater is assumed
// forward-compatible unless proven otherwise. Bump Max on every
// EMLyUpdater release, even one that doesn't touch this package; bump Min
// only when EMLy requires a newer protocolVersion.
const (
	MinCompatibleUpdaterVersion = "1.2.0b"
	MaxCompatibleUpdaterVersion = "1.2.0b"
)

// version returns this running EMLy build's own semver, read from its
// config.ini GUI_SEMVER — the same value the installer/updater writes on
// every update, so there is no separate hardcoded copy to keep in sync.
// Falls back to "0.0.0" (always below the server's MinCompatibleEMLyVersion
// floor) if config.ini can't be read, so a broken/missing config fails the
// server-side version gate loudly rather than silently skipping it.
func version() string {
	cfg, err := utils.LoadConfig(utils.DefaultConfigPath())
	if err != nil || cfg == nil || cfg.EMLy.GUISemver == "" {
		return "0.0.0"
	}
	return cfg.EMLy.GUISemver
}

// less reports whether semver string a is strictly older than b.
func less(a, b string) (bool, error) {
	va, err := goversion.NewVersion(a)
	if err != nil {
		return false, fmt.Errorf("invalid version %q: %w", a, err)
	}
	vb, err := goversion.NewVersion(b)
	if err != nil {
		return false, fmt.Errorf("invalid version %q: %w", b, err)
	}
	return va.LessThan(vb), nil
}

// checkPeerVersion enforces MinCompatibleUpdaterVersion against a
// responding EMLyUpdater's declared sender_version. Fails closed: a
// missing or unparseable sender_version is treated the same as one that
// is too old. A version above MaxCompatibleUpdaterVersion is accepted
// without complaint — see the doc comment on the consts above.
func checkPeerVersion(senderVersion string) error {
	if senderVersion == "" {
		return fmt.Errorf("updateripc: updater did not report its version (requires >= %s)", MinCompatibleUpdaterVersion)
	}
	belowMin, err := less(senderVersion, MinCompatibleUpdaterVersion)
	if err != nil {
		return fmt.Errorf("updateripc: invalid updater sender_version %q: %w", senderVersion, err)
	}
	if belowMin {
		return fmt.Errorf("updateripc: EMLyUpdater %s is older than the minimum supported %s", senderVersion, MinCompatibleUpdaterVersion)
	}
	return nil
}
