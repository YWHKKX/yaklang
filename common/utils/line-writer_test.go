package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// Appending via NewFileLineWriter must not synthesize a leading newline.
// Regression: the original code did `f.ReadAt(buf, size-1)` on an O_WRONLY
// handle, which always fails and leaves buf[0] == 0, and for an empty file it
// also seeks to offset -1. Both paths wrote a spurious "\n" before the first
// appended line.
func TestNewFileLineWriterAppend(t *testing.T) {
	tests := []struct {
		name string
		init string
		want string
	}{
		{"empty file", "", "b"},
		{"ends with newline", "a\n", "a\nb"},
		{"no trailing newline", "a", "a\nb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "payload.txt")
			if err := os.WriteFile(path, []byte(tt.init), 0o644); err != nil {
				t.Fatal(err)
			}

			w, err := NewFileLineWriter(path, os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			if _, err := w.WriteLineString("b"); err != nil {
				t.Fatalf("write: %v", err)
			}
			if err := w.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tt.want {
				t.Fatalf("got %q, want %q", string(got), tt.want)
			}
		})
	}
}
