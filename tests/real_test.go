package tests

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRealReleases runs whybumped on the go.mod and go.sum of real releases,
// fetched from GitHub, with the module proxy from the environment. It needs
// network access, so it only runs with WHYBUMPED_REAL=1 (make test-real).
func TestRealReleases(t *testing.T) {
	if os.Getenv("WHYBUMPED_REAL") != "1" {
		t.Skip("set WHYBUMPED_REAL=1 to fetch real releases")
	}
	cases := []struct {
		repo, old, new string
		code           int
		want           []string
	}{
		{"spf13/cobra", "v1.8.0", "v1.8.1", 0, []string{
			"6 modules before, 6 after, 5 unchanged",
			"INFO  raised github.com/cpuguy83/go-md2man/v2 v2.0.3 -> v2.0.4\n        go.mod requires this version and no dependency requires it",
			"0 warnings, 1 info",
		}},
		{"prometheus/alertmanager", "v0.27.0", "v0.28.0", 1, []string{
			"WARN  go-raised go 1.21 -> 1.22.0",
			"github.com/KimMachineGun/automemlimit@v0.7.0 needs go 1.22.0 and go.mod requires it directly",
			"WARN  lowered cloud.google.com/go/compute v1.20.1 -> v1.6.1",
			"golang.org/x/oauth2@v0.16.0 required v1.20.1 before",
			"requirers that moved: golang.org/x/oauth2@v0.16.0 (now v0.24.0)",
		}},
		{"caddyserver/caddy", "v2.8.4", "v2.9.0", 1, []string{
			"WARN  go-raised go 1.21.0 -> 1.22.3",
			"github.com/quic-go/quic-go@v0.48.2 needs go 1.22 and go.mod requires it directly",
			"go.mod went further than any dependency needs (go 1.22)",
			"INFO  toolchain-changed toolchain go1.22.2 -> go1.23.0",
			"WARN  lowered cloud.google.com/go/compute v1.24.0 -> v1.23.3",
			"google.golang.org/grpc@v1.63.2 required v1.24.0 before",
		}},
	}
	env := os.Environ()
	for _, c := range cases {
		t.Run(c.repo, func(t *testing.T) {
			old := fetchModule(t, c.repo, c.old)
			nw := fetchModule(t, c.repo, c.new)
			r := runEnv(t, repoRoot(t), env, "--old", old, "--new", nw)
			if r.code != c.code {
				t.Fatalf("exit %d, want %d\n%s%s", r.code, c.code, r.stdout, r.stderr)
			}
			mustContain(t, r.stdout, c.want...)
			r = runEnv(t, repoRoot(t), env, "--old", nw, "--new", nw)
			if r.code != 0 || !strings.Contains(r.stdout, "no dependency changes") {
				t.Errorf("%s against itself: exit %d\n%s%s", c.new, r.code, r.stdout, r.stderr)
			}
		})
	}
}

func fetchModule(t *testing.T, repo, ref string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), ref)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 30 * time.Second}
	for _, f := range []string{"go.mod", "go.sum"} {
		resp, err := client.Get("https://raw.githubusercontent.com/" + repo + "/" + ref + "/" + f)
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != 200 {
			t.Fatalf("fetch %s %s %s: %v %d", repo, ref, f, err, resp.StatusCode)
		}
		writeFile(t, filepath.Join(dir, f), string(body))
	}
	return dir
}
