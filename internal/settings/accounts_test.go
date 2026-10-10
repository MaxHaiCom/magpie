package settings

import (
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

func TestAccountDiscoveryOff(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	for _, v := range []string{"", "0", "false", "1", "true"} {
		t.Setenv("MAGPIE_NO_ACCOUNTS", v)
		if got, want := AccountDiscoveryOff(), v == "1" || v == "true"; got != want {
			t.Fatalf("%q: %v, want %v", v, got, want)
		}
	}
	s := Load()
	s.NoAccountDiscovery = true
	if err := Save(s); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAGPIE_NO_ACCOUNTS", "0")
	if !AccountDiscoveryOff() {
		t.Fatal("environment defeated saved opt-out")
	}
	from := Settings{}
	from.KeepOwn(s)
	if !from.NoAccountDiscovery {
		t.Fatal("sync lost this machine's opt-out")
	}
}
