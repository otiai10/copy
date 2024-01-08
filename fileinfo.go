package copy

import "io/fs"

// A FileInfo describes a file and is returned by Stat.
// This is a cloned definition of fs.FileInfo (go1.16~).
type fileInfo interface {
	// Name() string       // base name of the file
	// Size() int64        // length in bytes for regular files; system-dependent for others
	Mode() fs.FileMode // file mode bits
	// ModTime() time.Time // modification time
	IsDir() bool      // abbreviation for Mode().IsDir()
	Sys() interface{} // underlying data source (can return nil)
}
