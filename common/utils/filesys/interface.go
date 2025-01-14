package filesys

import (
	"io/fs"
	"strings"
)

// FileSystem defines the methods of an abstract filesystem.
type FileSystem interface {
	ReadFile(name string) ([]byte, error)

	Open(name string) (fs.File, error)
	// RelOpen(name string) (fs.File, error)

	// Stat returns a FileInfo describing the file.
	// If there is an error, it should be of type *PathError.
	Stat(name string) (fs.FileInfo, error)
	// RelStat(name string) (fs.FileInfo, error)
	// ReadDir reads the named directory
	// and returns a list of directory entries sorted by filename.
	ReadDir(name string) ([]fs.DirEntry, error)

	Join(elem ...string) string
	// Rel(targpath string) (string, error)

	GetSeparators() rune

	PathSplit(string) (string, string)
	Ext(string) string
}

func splitWithSeparator(path string, sep rune) (string, string) {
	if len(path) == 0 {
		return "", ""
	}
	idx := strings.LastIndexFunc(path, func(r rune) bool {
		if r == sep {
			return true
		}
		return false
	})
	if idx == -1 {
		return "", path
	}
	return path[:idx], path[idx:]
}

func getExtension(path string) string {
	if len(path) == 0 {
		return ""
	}
	idx := strings.LastIndexFunc(path, func(r rune) bool {
		if r == '.' {
			return true
		}
		return false
	})
	if idx == -1 {
		return ""
	}
	return path[idx:]
}
