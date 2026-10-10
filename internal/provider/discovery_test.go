package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/settings"
)

func TestAccountDiscoveryOffReadsNoCredentials(t *testing.T) {
	for _, mode := range []string{"setting", "environment"} {
		t.Run(mode, func(t *testing.T) {
			home := claudeHome(t)
			t.Setenv("MAGPIE_NO_ACCOUNTS", "")
			writeFile(t, filepath.Join(home, ".claude", ".credentials.json"), map[string]any{})
			writeFile(t, filepath.Join(home, ".codex", "auth.json"), map[string]any{
				"auth_mode": "chatgpt", "tokens": map[string]any{"access_token": "fixture-access", "account_id": "fixture-account"},
			})
			old := accountReadFile
			reads := 0
			accountReadFile = func(path string) ([]byte, error) { reads++; return os.ReadFile(path) }
			t.Cleanup(func() { accountReadFile = old; ForgetAccounts() })
			// The same fixture is discoverable by default, and the counter sees it.
			if _, ok := codexAccount(home); !ok || reads == 0 {
				t.Fatal("default discovery did not read the fixture")
			}
			reads = 0
			ForgetAccounts()
			if mode == "setting" {
				s := settings.Load()
				s.NoAccountDiscovery = true
				if err := settings.Save(s); err != nil {
					t.Fatal(err)
				}
			} else {
				t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
			}
			if got := Accounts(); len(got) != 0 {
				t.Fatalf("discovered accounts: %v", got)
			}
			if !loginsSeenAt.IsZero() {
				t.Fatal("Accounts remembered logins while disabled")
			}
			rememberLogins(true)
			if !loginsSeenAt.IsZero() {
				t.Fatal("forced remembering ran while disabled")
			}
			for _, agent := range append([]string{""}, builtinOrder...) {
				if got := Logins(agent); len(got) != 0 {
					t.Fatalf("%s logins: %v", agent, got)
				}
			}
			ctx := context.Background()
			if len(MoveCandidates()) != 0 || len(HandOver(ctx, true)) != 0 || len(MoveRetiring(ctx)) != 0 {
				t.Fatal("account migration ran")
			}
			if len(readMigrationsFile()) != 0 || len(loadStepFun()) != 0 {
				t.Fatal("saved subscription credentials read")
			}
			signed := false
			cached := Provider{Account: &Account{sign: func(context.Context, *http.Request, []byte) error { signed = true; return nil }}}
			if err := cached.Sign(ctx, httptest.NewRequest(http.MethodPost, "http://127.0.0.1:1", nil), Chat, nil); err == nil || signed {
				t.Fatal("cached account authenticated a request")
			}

			if got := SubscriptionUsage(ctx); len(got) != 0 {
				t.Fatal("subscription usage was read")
			}
			if got := fetchSubscriptionUsage(ctx); len(got) != 0 {
				t.Fatal("subscription refresh was read")
			}
			if got := LoginUsage(ctx, "codex"); len(got) != 0 {
				t.Fatal("login usage was read")
			}
			if got := RenewLogins(ctx, 0); len(got) != 0 {
				t.Fatal("accounts were renewed")
			}
			if got := SessionIdentities(filepath.Join(home, ".codex")); len(got) != 0 {
				t.Fatal("session identities were read")
			}
			if _, _, ok := readClaudeCredential(); ok {
				t.Fatal("Claude credential read")
			}
			if _, ok := codexAccount(home); ok {
				t.Fatal("Codex credential read")
			}
			if _, ok := liveLogin("codex"); ok {
				t.Fatal("live login read")
			}
			if CodexAPIKeySignedIn() {
				t.Fatal("Codex sign-in read")
			}
			if _, _, err := codexToken(ctx, codexAuthPath()); err == nil {
				t.Fatal("Codex token read")
			}
			if _, _, err := DevinAuthAt(""); err == nil {
				t.Fatal("Devin auth read")
			}
			if readCursorToken() != "" {
				t.Fatal("Cursor token read")
			}
			if _, ok := copilotCLILogin(); ok {
				t.Fatal("Copilot login read")
			}
			if reads != 0 {
				t.Fatalf("credential reads while disabled: %d", reads)
			}
			if _, err := os.Stat(loginsPath()); !os.IsNotExist(err) {
				t.Fatalf("remembered logins: %v", err)
			}
			// API-key providers still belong to the catalog.
			if err := Save(Provider{ID: "fixture-api", Key: "fixture-key", Chat: "http://127.0.0.1:1"}); err != nil {
				t.Fatal(err)
			}
			if got := All(); len(got) != 1 || got[0].Account != nil {
				t.Fatalf("API providers: %v", got)
			}
		})
	}
}

func TestAccountDiscoveryOffHidesCachedCredentials(t *testing.T) {
	home := claudeHome(t)
	t.Setenv("MAGPIE_NO_ACCOUNTS", "")
	claudeSignIn(t, home, time.Now().Add(time.Hour))
	if _, _, ok := claudeCredential(); !ok {
		t.Fatal("fixture not cached")
	}
	release := Hold()
	t.Cleanup(release)
	if len(All()) == 0 {
		t.Fatal("positive control did not hold accounts")
	}
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	if len(All()) != 0 {
		t.Fatal("held account exposed while disabled")
	}
	if _, _, ok := claudeCredential(); ok {
		t.Fatal("cached credential exposed while disabled")
	}
	t.Cleanup(ForgetAccounts)
}

func TestAccountDiscoveryOffStartsNoKeychainCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX keychain stand-in")
	}
	home := claudeHome(t)
	t.Setenv("MAGPIE_NO_ACCOUNTS", "")
	old := claudeKeychain
	claudeKeychain = true
	t.Cleanup(func() { claudeKeychain = old; ForgetAccounts() })
	bin := filepath.Join(home, "bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	marker := filepath.Join(home, "keychain-calls")
	t.Setenv("DISCOVERY_KEYCHAIN_MARKER", marker)
	if err := os.WriteFile(filepath.Join(bin, "security"), []byte("#!/bin/sh\nprintf 'called\\n' >> \"$DISCOVERY_KEYCHAIN_MARKER\"\nexit 1\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	readClaudeCredential()
	before, err := os.ReadFile(marker)
	if err != nil || len(before) == 0 {
		t.Fatalf("keychain positive control: %v", err)
	}
	t.Setenv("MAGPIE_NO_ACCOUNTS", "1")
	readClaudeCredential()
	readClaudeDir(home)
	cursorKeychainToken(true)
	after, err := os.ReadFile(marker)
	if err != nil || string(after) != string(before) {
		t.Fatal("keychain command started while discovery was off")
	}
}
