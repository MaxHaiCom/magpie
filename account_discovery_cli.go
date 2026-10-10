package main

import (
	"fmt"

	"github.com/yetone/magpie/internal/provider"
	"github.com/yetone/magpie/internal/settings"
)

// accountsDiscoveryCmd sets whether agents' sign-ins are discovered, before
// startup looks at them. With no argument it reports the effective preference.
func accountsDiscoveryCmd(args []string) error {
	if len(args) > 1 || len(args) == 1 && args[0] != "on" && args[0] != "off" {
		return fmt.Errorf("usage: magpie accounts discovery [on|off]")
	}
	if len(args) == 1 {
		s := settings.Load()
		s.NoAccountDiscovery = args[0] == "off"
		if err := settings.Save(s); err != nil {
			return err
		}
		provider.ForgetAccounts()
	}
	if settings.AccountDiscoveryOff() {
		fmt.Println("account discovery is off")
	} else {
		fmt.Println("account discovery is on")
	}
	return nil
}
