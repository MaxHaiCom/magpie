package plugin

import (
	"context"
	"os"
	"testing"

	"github.com/yetone/magpie/internal/settings"
)

func TestAccountDiscoveryOffKeepsPluginsIdle(t *testing.T) {
	for _, mode := range []string{"setting", "environment"} {
		t.Run(mode, func(t *testing.T) {
			storeSandbox(t)
			t.Setenv("MAGPIE_NO_ACCOUNTS", "")
			body := []byte(`{"fixture":{"type":"oauth","access":"fixture-access","refresh":"fixture-refresh"}}`)
			if err := os.WriteFile(AuthPath(), body, 0o600); err != nil {
				t.Fatal(err)
			}
			if len(Auths("fixture")) != 1 {
				t.Fatal("positive control did not find fixture")
			}
			if mode == "setting" {
				s := settings.Load()
				s.NoAccountDiscovery = true
				if err := settings.Save(s); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
			}
			if len(Auths("fixture")) != 0 || len(readAuth()) != 0 || len(listingProxies()) != 0 {
				t.Fatal("plugin credentials exposed")
			}
			if _, err := readAuthFile(); !os.IsNotExist(err) {
				t.Fatalf("credential reader was not blocked: %v", err)
			}
			if h, err := get(context.Background()); h != nil || err == nil {
				t.Fatal("plugin host was not blocked")
			}
			after, err := os.ReadFile(AuthPath())
			if err != nil || string(after) != string(body) {
				t.Fatal("saved sign-in changed")
			}
		})
	}
}
