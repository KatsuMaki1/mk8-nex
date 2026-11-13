package globals

import "os"

// NEX configuration for Mario Kart 8 (Wii U): game server ID 1010EB00,
// access key from the retail executable, NEX 3.5.4.
const (
	GameServerID = "1010eb00"
	AccessKey    = "25dbf96a"

	NEXMajor = 3
	NEXMinor = 5
	NEXPatch = 4
)

// Wire flags, env-tunable because NEX 3.5 sits on the legacy/modern boundary.
// Mario Kart 8 needs the modern CONNECT signature and the structure header.
func envFlag(name string, def bool) bool {
	switch os.Getenv(name) {
	case "1":
		return true
	case "0":
		return false
	}
	return def
}

func LegacySignature() bool { return envFlag("PN_MK8_LEGACY_SIG", false) }
func StructureHeader() bool { return envFlag("PN_MK8_STRUCT_HEADER", true) }
