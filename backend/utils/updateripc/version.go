package updateripc

import (
	"fmt"

	goversion "github.com/hashicorp/go-version"

	"emly/backend/utils"
)

// MinCompatibleUpdaterVersionV1 / MaxCompatibleUpdaterVersionV1 bound the
// EMLyUpdater releases known to speak the frozen legacy protocol_version 1
// one-shot exchange — see the compatibility matrix atop
// proto/updateripc.proto. Frozen: this client no longer speaks v1 at all
// (see handshake.go), these consts exist only to document the matrix row.
const (
	MinCompatibleUpdaterVersionV1 = "1.2.0b"
	MaxCompatibleUpdaterVersionV1 = "1.2.0b"
)

// MinCompatibleUpdaterVersionV2 / MaxCompatibleUpdaterVersionV2 bound the
// EMLyUpdater releases known to speak the v2 handshake. A ServerAnswHello
// whose server_version is below Min is rejected even though the wire
// protocol matched, since not every required fix changes the wire format.
// Max is informational only (never enforced): a newer EMLyUpdater is
// assumed forward-compatible unless proven otherwise. Bump Max on every
// EMLyUpdater release, even one that doesn't touch this package; bump Min
// only when EMLy requires a newer protocolVersionV2.
const (
	MinCompatibleUpdaterVersionV2 = "1.3.0"
	MaxCompatibleUpdaterVersionV2 = "1.3.0"
)

// version returns this running EMLy build's own semver, read from its
// config.ini GUI_SEMVER — the same value the installer/updater writes on
// every update, so there is no separate hardcoded copy to keep in sync.
// Falls back to "0.0.0" (always below the server's MinCompatibleEMLyVersionV2
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

// checkPeerVersion enforces min against senderVersion. Fails closed: a
// missing or unparseable senderVersion is treated the same as one that is
// too old. A version above max is accepted without complaint.
func checkPeerVersion(senderVersion, min, max string) error {
	if senderVersion == "" {
		return fmt.Errorf("updateripc: updater did not report its version (requires >= %s)", min)
	}
	belowMin, err := less(senderVersion, min)
	if err != nil {
		return fmt.Errorf("updateripc: invalid updater sender_version %q: %w", senderVersion, err)
	}
	if belowMin {
		return fmt.Errorf("updateripc: EMLyUpdater %s is older than the minimum supported %s", senderVersion, min)
	}
	return nil
}

// checkPeerVersionV1 enforces the frozen legacy (protocol_version 1)
// compatibility range. Unused by any current code path (this client only
// speaks v2 — see handshake.go) — kept so the frozen matrix row has a
// documented, testable definition rather than a bare comment.
func checkPeerVersionV1(senderVersion string) error {
	return checkPeerVersion(senderVersion, MinCompatibleUpdaterVersionV1, MaxCompatibleUpdaterVersionV1)
}

// checkPeerVersionV2 enforces the v2 handshake compatibility range against
// a ServerAnswHello's server_version.
func checkPeerVersionV2(senderVersion string) error {
	return checkPeerVersion(senderVersion, MinCompatibleUpdaterVersionV2, MaxCompatibleUpdaterVersionV2)
}
