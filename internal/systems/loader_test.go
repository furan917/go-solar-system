package systems

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func systemJSON(displayName string) string {
	return describedSystemJSON(displayName, "d")
}

func describedSystemJSON(displayName, description string) string {
	return `{"systemName":"` + displayName + `","description":"` + description + `","bodies":[{"englishName":"b"}]}`
}

func filterTestManager(t *testing.T) *SystemManager {
	t.Helper()

	bundled := fstest.MapFS{
		"kepler-452.json":     {Data: []byte(describedSystemJSON("Kepler-452", "Home of Earth's cousin"))},
		"trappist-1.json":     {Data: []byte(describedSystemJSON("TRAPPIST-1", "Seven Earth-sized planets"))},
		"alpha-centauri.json": {Data: []byte(describedSystemJSON("Alpha Centauri", "Closest binary star system"))},
	}

	sm := NewSystemManager(bundled)
	if err := sm.ScanSystems(); err != nil {
		t.Fatalf("ScanSystems: %v", err)
	}

	return sm
}

func filterNames(t *testing.T, sm *SystemManager, query string) []string {
	t.Helper()

	var names []string
	for _, entry := range sm.FilterSystems(query) {
		names = append(names, entry.Name)
	}

	return names
}

func TestFilterSystemsEmptyQueryReturnsAll(t *testing.T) {
	sm := filterTestManager(t)

	if got := sm.FilterSystems(""); len(got) != 4 {
		t.Fatalf("empty query should return every system, got %d", len(got))
	}
	if got := sm.FilterSystems("   "); len(got) != 4 {
		t.Fatalf("whitespace query should return every system, got %d", len(got))
	}
}

func TestFilterSystemsMatchesNameCaseInsensitively(t *testing.T) {
	sm := filterTestManager(t)

	got := filterNames(t, sm, "KEPLER")
	if len(got) != 1 || got[0] != "kepler-452" {
		t.Errorf("got %v, want [kepler-452]", got)
	}
}

func TestFilterSystemsMatchesDescription(t *testing.T) {
	sm := filterTestManager(t)

	got := filterNames(t, sm, "binary")
	if len(got) != 1 || got[0] != "alpha-centauri" {
		t.Errorf("got %v, want [alpha-centauri]", got)
	}
}

func TestFilterSystemsRequiresEveryTerm(t *testing.T) {
	sm := filterTestManager(t)

	if got := filterNames(t, sm, "seven earth"); len(got) != 1 || got[0] != "trappist-1" {
		t.Errorf("terms should be ANDed, got %v", got)
	}
	if got := filterNames(t, sm, "seven nonsense"); len(got) != 0 {
		t.Errorf("unmatched term should exclude the entry, got %v", got)
	}
}

func TestFilterSystemsNoMatches(t *testing.T) {
	sm := filterTestManager(t)

	if got := sm.FilterSystems("zzzz"); len(got) != 0 {
		t.Errorf("expected no matches, got %v", got)
	}
}

func TestListSystemsPairsNameWithInfo(t *testing.T) {
	sm := filterTestManager(t)

	entries := sm.ListSystems()
	infoLines, err := sm.ListSystemsWithInfo()
	if err != nil {
		t.Fatalf("ListSystemsWithInfo: %v", err)
	}

	if len(entries) != len(infoLines) {
		t.Fatalf("entry and info counts diverged: %d vs %d", len(entries), len(infoLines))
	}
	for i := range entries {
		if entries[i].Info != infoLines[i] {
			t.Errorf("index %d: %q != %q", i, entries[i].Info, infoLines[i])
		}
	}
}

func newTestManager(t *testing.T, userDir string) *SystemManager {
	t.Helper()

	bundled := fstest.MapFS{
		"alpha-centauri.json": {Data: []byte(systemJSON("Alpha Centauri"))},
		"README.md":           {Data: []byte("not a system")},
	}

	sm := NewSystemManager(bundled, userDir)
	if err := sm.ScanSystems(); err != nil {
		t.Fatalf("ScanSystems: %v", err)
	}

	return sm
}

func TestScanSystemsFindsBundledSystems(t *testing.T) {
	sm := newTestManager(t, "")

	got := sm.GetAvailableSystems()
	want := []string{"solar-system", "alpha-centauri"}

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestScanSystemsIgnoresMissingUserDir(t *testing.T) {
	sm := newTestManager(t, filepath.Join(t.TempDir(), "does-not-exist"))

	if len(sm.GetAvailableSystems()) != 2 {
		t.Fatalf("missing user dir should be ignored, got %v", sm.GetAvailableSystems())
	}
}

func TestUserSystemsAreAddedAndShadowBundled(t *testing.T) {
	dir := t.TempDir()
	write := func(name, contents string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	write("my-system.json", systemJSON("My System"))
	write("alpha-centauri.json", systemJSON("Overridden Centauri"))

	sm := newTestManager(t, dir)

	if len(sm.GetAvailableSystems()) != 3 {
		t.Fatalf("expected user system to be added, got %v", sm.GetAvailableSystems())
	}

	data, err := sm.LoadSystem("my-system")
	if err != nil {
		t.Fatalf("LoadSystem(my-system): %v", err)
	}
	if data.SystemName != "My System" {
		t.Errorf("got %q, want %q", data.SystemName, "My System")
	}

	overridden, err := sm.LoadSystem("alpha-centauri")
	if err != nil {
		t.Fatalf("LoadSystem(alpha-centauri): %v", err)
	}
	if overridden.SystemName != "Overridden Centauri" {
		t.Errorf("user file should shadow bundled, got %q", overridden.SystemName)
	}
}

func TestSystemsSearchDirsIncludesConfigDirAndEnvDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	t.Setenv(SystemsDirEnvVar, "/custom/systems")

	got := SystemsSearchDirs()
	if len(got) != 3 {
		t.Fatalf("env dir should be added, not replace the config dir: %v", got)
	}
	if got[0] != LocalSystemsDir {
		t.Errorf("local systems dir should still be scanned, got %v", got)
	}
	if got[2] != "/custom/systems" {
		t.Errorf("env dir should take precedence, got %v", got)
	}
}

func TestSystemsSearchDirsWithoutEnvVar(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/xdg")
	t.Setenv(SystemsDirEnvVar, "")

	if got := SystemsSearchDirs(); len(got) != 2 {
		t.Errorf("expected the local and config dirs, got %v", got)
	}
}

func TestBothUserDirsAreScanned(t *testing.T) {
	configDir, envDir := t.TempDir(), t.TempDir()

	if err := os.WriteFile(filepath.Join(configDir, "from-config.json"), []byte(systemJSON("From Config")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(envDir, "from-env.json"), []byte(systemJSON("From Env")), 0o644); err != nil {
		t.Fatal(err)
	}

	bundled := fstest.MapFS{"alpha-centauri.json": {Data: []byte(systemJSON("Alpha Centauri"))}}
	sm := NewSystemManager(bundled, configDir, envDir)
	if err := sm.ScanSystems(); err != nil {
		t.Fatalf("ScanSystems: %v", err)
	}

	if len(sm.GetAvailableSystems()) != 4 {
		t.Fatalf("expected systems from both user dirs, got %v", sm.GetAvailableSystems())
	}
}
