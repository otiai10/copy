package copy

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func preserveTime(info os.FileInfo, dest string) error {
	mtime := unix.NsecToTimeval(info.ModTime().UnixNano())

	statPtr := info.Sys()
	if statPtr == nil {
		return fmt.Errorf("Error in obtain stat structure")
	}
	stat := statPtr.(*syscall.Stat_t)
	atime := unix.NsecToTimeval(int64(stat.Atim.Sec)*1e9 + int64(stat.Atim.Nsec))
	return unix.Lutimes(dest, []unix.Timeval{atime, mtime})
}
