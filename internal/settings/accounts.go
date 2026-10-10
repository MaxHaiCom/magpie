package settings

import "os"

// AccountDiscoveryOff says not to read agents' sign-ins, including keychains.
// The environment can turn discovery off without changing the saved preference.
func AccountDiscoveryOff() bool {
	if v := os.Getenv("MAGPIE_NO_ACCOUNTS"); v != "" && v != "0" && v != "false" {
		return true
	}
	return Load().NoAccountDiscovery
}
