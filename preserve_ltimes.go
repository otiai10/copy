//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package copy

import (
	"os"

	"golang.org/x/sys/unix"
)

func preserveLtimes(srcinfo os.FileInfo, dest string) error {
	spec := getTimeSpec(srcinfo)

	return unix.Lutimes(dest, []unix.Timeval{
		unix.NsecToTimeval(spec.Atime.UnixNano()),
		unix.NsecToTimeval(spec.Mtime.UnixNano()),
	})
}
