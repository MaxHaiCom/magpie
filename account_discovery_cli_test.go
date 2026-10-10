package main

import (
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/settings"
	"github.com/yetone/magpie/internal/testenv"
)

func TestAccountsDiscoveryCommand(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("MAGPIE_NO_ACCOUNTS", "")
	for _, arg := range []string{"off", "on"} {
		if err := run([]string{"accounts", "discovery", arg}); err != nil {
			t.Fatal(err)
		}
		if settings.Load().NoAccountDiscovery != (arg == "off") {
			t.Fatalf("%s not saved", arg)
		}
	}
	if err := accountsDiscoveryCmd([]string{"typo"}); err == nil {
		t.Fatal("invalid choice accepted")
	}
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	if err := accountsDiscoveryCmd([]string{"on"}); err != nil {
		t.Fatal(err)
	}
	if !settings.AccountDiscoveryOff() {
		t.Fatal("command defeated environment opt-out")
	}
}
