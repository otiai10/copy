//go:build linux

package copy

import (
	"fmt"
	"os"

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
// Linux implementation uses the `ficlone` ioctl:
// https://manpages.debian.org/testing/manpages-dev/ioctl_ficlone.2.en.html
//
// Support:
//   - BTRFS or XFS filesystem
//
// Considerations:
//   - Ownership is not preserved.
//   - Setuid and Setgid are not preserved.
//   - Times are not preserved.
var ReflinkCopy = FileCopyMethod{
	fcopy: func(src, dest string, info os.FileInfo, opt Options) (err error, skipFile bool) {
		if opt.FS != nil {
			return fmt.Errorf("%w: cannot create reflink from Go's fs.FS interface", ErrUnsupportedCopyMethod), false
		}

		if opt.WrapReader != nil {
			return fmt.Errorf("%w: cannot create reflink when WrapReader option is used", ErrUnsupportedCopyMethod), false
		}

		// Open source file.
		readcloser, err := os.OpenFile(src, os.O_RDONLY, 0)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, true
			}
			return
		}
		defer fclose(readcloser, &err)

		// Open dest file.
		f, err := os.Create(dest)
		if err != nil {
			return
		}
		defer fclose(f, &err)

		// Do copy.
		srcFd := readcloser.Fd()
		destFd := f.Fd()
		err = unix.IoctlFileClone(int(destFd), int(srcFd))

		// Return an error if cloning is not possible.
		if err != nil {
			_ = os.Remove(dest) // remove the empty file on error
			return &os.PathError{
				Op:   "create reflink",
				Path: src,
				Err:  err,
			}, false
		}

		return nil, false
	},
}
