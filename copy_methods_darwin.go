//go:build darwin

package copy

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
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
//
// -------------------- PLATFORM SPECIFIC INFORMATION --------------------
//
// Darwin implementation uses the `clonefile` syscall:
// https://www.manpagez.com/man/2/clonefile/
//
// Support:
//   - MacOS 10.14 or newer
//   - APFS filesystem
//
// Considerations:
//   - Ownership is not preserved.
//   - Setuid and Setgid are not preserved.
//   - Times are copied by default.
//   - Flag CLONE_NOFOLLOW is not used, we use lcopy instead of fcopy for
//     symbolic links.
var ReflinkCopy = FileCopyMethod{
	fcopy: func(src, dest string, info os.FileInfo, opt Options) (err error, skipFile bool) {
		if opt.FS != nil {
			return fmt.Errorf("%w: cannot create reflink from Go's fs.FS interface", ErrUnsupportedCopyMethod), false
		}

		if opt.WrapReader != nil {
			return fmt.Errorf("%w: cannot create reflink when WrapReader option is used", ErrUnsupportedCopyMethod), false
		}

		// Do copy.
		const clonefileFlags = 0
		err = unix.Clonefile(src, dest, clonefileFlags)

		// If the error is the file already exists, delete it and try again.
		if errors.Is(err, os.ErrExist) {
			if err = os.Remove(dest); err != nil {
				return err, false
			}

			err = unix.Clonefile(src, dest, clonefileFlags) // retry
		}

		// Return error if clone is not possible.
		if err != nil {
			if os.IsNotExist(err) {
				return nil, true // but not if source file doesn't exist
			}

			return &os.PathError{
				Op:   "create reflink",
				Path: src,
				Err:  err,
			}, false
		}

		// Copy-on-write preserves the modtime by default.
		// If PreserveTimes is not true, update the time to now.
		if !opt.PreserveTimes {
			now := time.Now()
			if err := os.Chtimes(dest, now, now); err != nil {
				return err, false
			}
		}

		return nil, false
	},
}
