package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDatabasePathsUsesDataDirEnvironment(t *testing.T) {
	dataDir := filepath.Join(t.TempDir(), "relay-data")
	t.Setenv(dataDirEnvVar, dataDir)

	paths, err := resolveDatabasePaths()
	if err != nil {
		t.Fatalf("resolveDatabasePaths() returned an unexpected error: %v", err)
	}

	expectedDataDir, errAbs := filepath.Abs(dataDir)
	if errAbs != nil {
		t.Fatalf("filepath.Abs() returned an unexpected error: %v", errAbs)
	}

	if paths.dataDir != expectedDataDir {
		t.Fatalf("dataDir = %q; expected %q", paths.dataDir, expectedDataDir)
	}

	if paths.yata != filepath.Join(expectedDataDir, "yata.db") {
		t.Fatalf("yata path = %q; expected %q", paths.yata, filepath.Join(expectedDataDir, "yata.db"))
	}

	if paths.catalog != filepath.Join(expectedDataDir, "catalog.db") {
		t.Fatalf("catalog path = %q; expected %q", paths.catalog, filepath.Join(expectedDataDir, "catalog.db"))
	}

	info, errStat := os.Stat(expectedDataDir)
	if errStat != nil {
		t.Fatalf("data directory was not created: %v", errStat)
	}

	if !info.IsDir() {
		t.Fatalf("data path = %q is not a directory", expectedDataDir)
	}
}

func TestResolveDatabasePathsUsesDefaultDataDirectoryWhenDataDirIsEmpty(t *testing.T) {
	t.Setenv(dataDirEnvVar, " ")

	paths, err := resolveDatabasePaths()
	if err != nil {
		t.Fatalf("resolveDatabasePaths() returned an unexpected error: %v", err)
	}

	if filepath.Base(paths.dataDir) != "data" {
		t.Fatalf("default dataDir = %q; expected a data directory", paths.dataDir)
	}

	if filepath.Dir(paths.yata) != paths.dataDir {
		t.Fatalf("yata path = %q; expected it inside %q", paths.yata, paths.dataDir)
	}

	if filepath.Dir(paths.catalog) != paths.dataDir {
		t.Fatalf("catalog path = %q; expected it inside %q", paths.catalog, paths.dataDir)
	}
}
