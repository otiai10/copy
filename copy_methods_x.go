//go:build !darwin

package copy

import (
	"os"
)

// ReflinkCopy tries to copy the file by creating a reflink from the source
// file to the destination file. This asks the filesystem to share the
// contents between the files using a copy-on-write method.
//
// Reflinks are the fastest way to copy large files, but have a few limitations:
//
//   - Requires using a supported filesystem (btrfs, xfs, apfs)
//   - Source and destination must be on the same filesystem.
//
// See: https://btrfs.readthedocs.io/en/latest/Reflink.html
var ReflinkCopy = FileCopyMethod{
	fcopy: func(src, dest string, info os.FileInfo, opt Options) (err error, skipFile bool) {
		// Not supported os.
		return ErrUnsupportedCopyMethod, false
	},
}
