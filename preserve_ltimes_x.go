//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

package copy

import (
	"os"

	"golang.org/x/sys/unix"
)

func preserveLtimes(srcinfo os.FileInfo, dest string) error {
	spec := getTimeSpec(srcinfo)

	if err := unix.Lutimes(dest, []unix.Timeval{
		{Sec: spec.Atime.Unix(), Usec: spec.Atime.UnixNano() / 1000 % 1000},
		{Sec: spec.Mtime.Unix(), Usec: spec.Mtime.UnixNano() / 1000 % 1000},
	}); err != nil {
		return err
	}
	return nil
}
