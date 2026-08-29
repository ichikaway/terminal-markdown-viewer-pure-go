package pager

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestHasLessUsesPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("executable fixture is Unix-specific")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "less")
	if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	if !HasLess() {
		t.Fatal("less was not found on PATH")
	}
}
