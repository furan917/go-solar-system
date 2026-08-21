package systems

import (
	"io/fs"
	"os"
	"path/filepath"
)

// SystemsDirEnvVar names an additional directory to load user systems from
const SystemsDirEnvVar = "GO_SOLAR_SYSTEM_SYSTEMS_DIR"

// LocalSystemsDir is the working-directory-relative systems path
const LocalSystemsDir = "systems"

// systemSource records where a system definition was found
type systemSource struct {
	fsys fs.FS
	path string
}

// SystemsSearchDirs returns the directories scanned for system files, in increasing order of precedence
func SystemsSearchDirs() []string {
	dirs := []string{LocalSystemsDir}

	if base, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(base, "go-solar-system", "systems"))
	}

	if dir := os.Getenv(SystemsDirEnvVar); dir != "" {
		dirs = append(dirs, dir)
	}

	return dirs
}
