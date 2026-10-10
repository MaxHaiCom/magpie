package agent

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yetone/magpie/internal/testenv"
)

func TestAccountDiscoveryOffSkipsAgentAuth(t *testing.T) {
	home := t.TempDir()
	testenv.SetHome(t, home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("MAGPIE_NO_ACCOUNTS", "")
	dir := filepath.Join(home, ".codex")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "auth.json")
	if err := os.WriteFile(path, []byte(`{"tokens":{"access_token":"fixture-access"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if !codexChatGPT(dir) {
		t.Fatal("positive control did not read fixture")
	}
	if _, err := readOwnAuth(path); err != nil {
		t.Fatal(err)
	}
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	if codexChatGPT(dir) {
		t.Fatal("Codex credential read while discovery was off")
	}
	if _, err := readOwnAuth(path); !os.IsNotExist(err) {
		t.Fatalf("agent credential reader not blocked: %v", err)
	}
}
