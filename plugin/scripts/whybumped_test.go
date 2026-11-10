package main

import (
	"archive/tar"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompareSemver(t *testing.T) {
	ordered := []string{
		"v0.0.0-20161208181325-20d25e280405",
		"v0.0.0-20190111050920-041ab4dc3f9d",
		"v0.1.0",
		"v1.0.0-alpha",
		"v1.0.0-alpha.1",
		"v1.0.0-alpha.beta",
		"v1.0.0-beta.2",
		"v1.0.0-beta.11",
		"v1.0.0-rc.1",
		"v1.0.0",
		"v1.0.1-0.20240101000000-abcdefabcdef",
		"v1.0.1",
		"v1.2.0",
		"v1.10.0",
		"v2.0.0+incompatible",
		"v10.0.0",
	}
	for i := range ordered {
		for j := range ordered {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := compareSemver(ordered[i], ordered[j]); got != want {
				t.Errorf("compareSemver(%s, %s) = %d, want %d", ordered[i], ordered[j], got, want)
			}
		}
	}
}

func TestParseSemverRejects(t *testing.T) {
	for _, v := range []string{"", "1.2.3", "v1", "v1.2", "v01.2.3", "v1.2.3-", "v1.2.3+", "v1.2.3-01", "v1.2.x", "latest", "v1.2.3 ", "v1.2.3-a..b"} {
		if _, ok := parseSemver(v); ok {
			t.Errorf("parseSemver(%q) accepted", v)
		}
	}
}

func TestSameRelease(t *testing.T) {
	for _, c := range []struct {
		a, b string
		want bool
	}{
		{"1.21", "1.21.0", true}, {"1.21.0", "1.21", true}, {"1.21", "1.21", false},
		{"1.21", "1.21.1", false}, {"1.21", "1.21rc1", false}, {"1.21", "1.22.0", false}, {"x", "1.21.0", false},
	} {
		if got := sameRelease(c.a, c.b); got != c.want {
			t.Errorf("sameRelease(%s, %s) = %v", c.a, c.b, got)
		}
	}
}

func TestCompareGo(t *testing.T) {
	ordered := []string{"1.9", "1.20", "1.21", "1.21beta1", "1.21rc1", "1.21rc2", "1.21.0", "1.21.8", "1.22", "1.22.0", "1.22.3", "2.0"}
	for i := range ordered {
		for j := range ordered {
			want := 0
			if i < j {
				want = -1
			} else if i > j {
				want = 1
			}
			if got := compareGo(ordered[i], ordered[j]); got != want {
				t.Errorf("compareGo(%s, %s) = %d, want %d", ordered[i], ordered[j], got, want)
			}
		}
	}
	for _, bad := range []string{"", "1", "go1.21", "1.x", "1.21.x", "1.21rc", "01.21", "1.021"} {
		if _, ok := parseGoVersion(bad); ok {
			t.Errorf("parseGoVersion(%q) accepted", bad)
		}
	}
}

const goodGraph = `example.com/m example.com/a@v1.1.0
example.com/m example.com/b@v1.0.0
example.com/m go@1.22
example.com/a@v1.1.0 example.com/b@v1.2.0
example.com/a@v1.1.0 go@1.21
example.com/b@v1.0.0 example.com/c@v0.1.0
example.com/b@v1.2.0 example.com/c@v0.3.0
go@1.22 toolchain@go1.22
`

func goodSide(t *testing.T) *side {
	t.Helper()
	s := &side{Main: "example.com/m", Selected: map[string]module{}, Edges: map[string][]string{}, Requirers: map[string][]string{}}
	if err := s.parseGraph([]byte(goodGraph)); err != nil {
		t.Fatal(err)
	}
	s.Selected["example.com/a"] = module{Path: "example.com/a", Version: "v1.1.0"}
	s.Selected["example.com/b"] = module{Path: "example.com/b", Version: "v1.2.0"}
	s.Selected["example.com/c"] = module{Path: "example.com/c", Version: "v0.3.0"}
	return s
}

func TestParseGraphAndCrossCheck(t *testing.T) {
	s := goodSide(t)
	if err := s.crossCheck(); err != nil {
		t.Fatal(err)
	}
	// go list says b is at v1.0.0: the graph disagrees.
	s.Selected["example.com/b"] = module{Path: "example.com/b", Version: "v1.0.0"}
	if err := s.crossCheck(); err == nil {
		t.Error("crossCheck accepted a version the graph doesn't select")
	}
	s = goodSide(t)
	s.Selected["example.com/z"] = module{Path: "example.com/z", Version: "v1.0.0"}
	if err := s.crossCheck(); err == nil {
		t.Error("crossCheck accepted a module the graph doesn't reach")
	}
	s = goodSide(t)
	delete(s.Selected, "example.com/c")
	if err := s.crossCheck(); err == nil {
		t.Error("crossCheck accepted a graph module that go list doesn't list")
	}
}

func TestParseGraphRejects(t *testing.T) {
	cases := map[string]string{
		"empty":             "",
		"blank":             "\n\n",
		"no newline":        strings.TrimSuffix(goodGraph, "\n"),
		"cut mid line":      goodGraph[:60],
		"one field":         "example.com/m\n",
		"three fields":      "example.com/m example.com/a@v1.0.0 x\n",
		"bad version":       "example.com/m example.com/a@1.0.0\n",
		"no version":        "example.com/m example.com/a\n",
		"bad go node":       "example.com/m go@banana\n",
		"go to module":      "go@1.22 example.com/a@v1.0.0\n",
		"build log":         "# example.com/m\n./main.go:3:2: undefined: x\n",
		"go error":          "go: example.com/a@v1.0.0: missing go.sum entry for go.mod file; to add it:\n",
		"test fail":         "--- FAIL: TestX (0.00s)\nFAIL\n",
		"binary":            "\x00\x01\x02\xff\n",
		"two mains":         "example.com/m example.com/a@v1.0.0\nexample.com/n example.com/a@v1.0.0\n",
		"quoted path":       "example.com/m \"example.com/a\"@v1.0.0\n",
		"main without deps": "example.com/m go@1.22\n" + "garbage line here\n",
	}
	for name, in := range cases {
		s := &side{Edges: map[string][]string{}, Requirers: map[string][]string{}}
		if err := s.parseGraph([]byte(in)); err == nil {
			t.Errorf("%s: parseGraph accepted %q", name, in)
		}
	}
}

func TestParseModulesRejects(t *testing.T) {
	cases := map[string]string{
		"empty":      "",
		"garbage":    "not json\n",
		"cut off":    `{"Path": "example.com/m", "Main": true}` + "\n" + `{"Path": "example.com/a", "Vers`,
		"error":      `{"Path": "example.com/a", "Version": "v1.0.0", "Error": {"Err": "module lookup disabled"}}`,
		"no version": `{"Path": "example.com/a"}`,
		"no path":    `{"Version": "v1.0.0"}`,
	}
	for name, in := range cases {
		if _, err := parseModules([]byte(in)); err == nil {
			t.Errorf("%s: parseModules accepted %q", name, in)
		}
	}
}

func TestChainFor(t *testing.T) {
	s := goodSide(t)
	got := strings.Join(chainFor(s, "example.com/c@v0.3.0"), " ")
	want := "example.com/c@v0.3.0 example.com/b@v1.2.0 example.com/a@v1.1.0"
	if got != want {
		t.Errorf("chain = %s, want %s", got, want)
	}
	if got := chainFor(s, "example.com/a@v1.1.0"); len(got) != 1 {
		t.Errorf("a direct requirement got chain %v", got)
	}
	// A cycle with no root still returns a chain instead of looping.
	s.Edges["example.com/c@v0.3.0"] = append(s.Edges["example.com/c@v0.3.0"], "example.com/a@v1.1.0")
	s.Requirers["example.com/a@v1.1.0"] = []string{"example.com/c@v0.3.0"}
	if got := chainFor(s, "example.com/c@v0.3.0"); len(got) < 2 {
		t.Errorf("cycle chain = %v", got)
	}
}

func TestExplainUpAndDown(t *testing.T) {
	n := goodSide(t)
	f := explainUp(n, "example.com/c", "v0.1.0", n.Selected["example.com/c"], "raised")
	if f.Severity != sevWarn || f.Cause != "pulled" || f.Chain[1] != "example.com/b@v1.2.0" {
		t.Errorf("explainUp c = %+v", f)
	}
	f = explainUp(n, "example.com/a", "v1.0.0", n.Selected["example.com/a"], "raised")
	if f.Severity != sevInfo || f.Cause != "requested" {
		t.Errorf("explainUp a = %+v", f)
	}
	// b@v1.0.0 isn't selected, but it is still in the graph and requires c.
	f = explainDown(n, n, "example.com/c", "v0.1.0", "v0.0.1")
	if f.Cause != "released" || f.Chain[1] != "example.com/b@v1.0.0" {
		t.Errorf("explainDown = %+v", f)
	}
}

func TestParseExpects(t *testing.T) {
	if _, err := parseExpects([]string{"example.com/a", "example.com/b@v1.2.3"}); err != nil {
		t.Error(err)
	}
	for _, bad := range []string{"", "@v1.0.0", "example.com/a@1.2.3", "example.com/a@latest", "a b"} {
		if _, err := parseExpects([]string{bad}); err == nil {
			t.Errorf("parseExpects(%q) accepted", bad)
		}
	}
}

func TestCheckGoMod(t *testing.T) {
	cases := map[string]string{
		"empty":       "",
		"spaces":      "  \n\n",
		"no newline":  "module example.com/m\n\ngo 1.22",
		"no module":   "go 1.22\n",
		"no go":       "module example.com/m\n",
		"binary":      "module example.com/m\x00\ngo 1.22\n",
		"not utf8":    "module example.com/m\n\xff\xfe\ngo 1.22\n",
		"build log":   "# example.com/m\n./main.go:3:2: undefined: x\n",
		"test output": "--- FAIL: TestX (0.00s)\nFAIL\nFAIL\texample.com/m\t0.01s\n",
	}
	for name, content := range cases {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, _, _, err := checkGoMod(dir); err == nil {
			t.Errorf("%s: checkGoMod accepted %q", name, content)
		}
	}
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/m // comment\n\ngo 1.22.0\n\ntoolchain go1.23.1\n"), 0o644)
	m, g, tc, err := checkGoMod(dir)
	if err != nil || m != "example.com/m" || g != "1.22.0" || tc != "go1.23.1" {
		t.Errorf("checkGoMod = %q %q %q %v", m, g, tc, err)
	}
	if _, _, _, err := checkGoMod(filepath.Join(dir, "go.mod")); err == nil {
		t.Error("checkGoMod accepted a file as directory")
	}
	if _, _, _, err := checkGoMod(filepath.Join(dir, "missing")); err == nil {
		t.Error("checkGoMod accepted a missing directory")
	}
}

func makeTar(t *testing.T, entries map[string]string, typ byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, body := range entries {
		h := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: typ}
		if typ == tar.TypeSymlink {
			h.Size = 0
			h.Linkname = body
		}
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if typ == tar.TypeReg {
			tw.Write([]byte(body))
		}
	}
	tw.Close()
	return buf.Bytes()
}

func TestUntar(t *testing.T) {
	dest := t.TempDir()
	if err := untar(makeTar(t, map[string]string{"sub/go.mod": "module x\n"}, tar.TypeReg), dest); err != nil {
		t.Fatal(err)
	}
	if b, err := os.ReadFile(filepath.Join(dest, "sub", "go.mod")); err != nil || string(b) != "module x\n" {
		t.Errorf("extracted %q %v", b, err)
	}
	for _, name := range []string{"../evil", "/abs/evil", "a/../../evil"} {
		if err := untar(makeTar(t, map[string]string{name: "x"}, tar.TypeReg), t.TempDir()); err == nil {
			t.Errorf("untar accepted %q", name)
		}
	}
	if err := untar(makeTar(t, map[string]string{"go.mod": "/etc/passwd"}, tar.TypeSymlink), dest); err == nil {
		t.Error("untar of only a symlink should report no files")
	}
	full := makeTar(t, map[string]string{"go.mod": strings.Repeat("x", 2000)}, tar.TypeReg)
	if err := untar(full[:700], t.TempDir()); err == nil {
		t.Error("untar accepted a cut-off archive")
	}
	if err := untar([]byte("garbage"), t.TempDir()); err == nil {
		t.Error("untar accepted garbage")
	}
}

func TestRunFlagErrors(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--old", "a", "--base", "HEAD"},
		{"--old", "a", "--format", "xml"},
		{"--old", "a", "--fail-on", "info"},
		{"--old", "a", "extra"},
		{"--old", "a", "--expect", "x@latest"},
		{"--nope"},
	} {
		var out, errb bytes.Buffer
		if code := run(args, &out, &errb); code != exitUntrusted {
			t.Errorf("run(%v) = %d, want %d", args, code, exitUntrusted)
		}
		if out.Len() != 0 {
			t.Errorf("run(%v) printed on stdout: %q", args, out.String())
		}
	}
	var out bytes.Buffer
	if code := run([]string{"--version"}, &out, &out); code != 0 || !strings.Contains(out.String(), version) {
		t.Errorf("--version: %d %q", code, out.String())
	}
}
