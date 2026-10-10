package provider

import (
	"errors"
	"os"

	"github.com/yetone/magpie/internal/settings"
)

var errAccountDiscoveryOff = errors.New("account discovery is off; magpie accounts discovery on enables subscriptions")

// accountReadFile is os.ReadFile; tests count reads of fixture credentials.
var accountReadFile = os.ReadFile

// readAccountFile keeps direct credential readers behind the same opt-out as Accounts.
func readAccountFile(path string) ([]byte, error) {
	if settings.AccountDiscoveryOff() {
		return nil, errAccountDiscoveryOff
	}
	return accountReadFile(path)
}
