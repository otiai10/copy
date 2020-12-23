// +build !windows,!darwin,!freebsd,386

// TODO: add more runtimes

package copy

import (
	"os"
	"syscall"
	"time"
)

func getTimeSpec(info os.FileInfo) timespec {
	stat := info.Sys().(*syscall.Stat_t)
	times := timespec{
		Mtime: info.ModTime(),
		Atime: time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec)),
		Ctime: time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec)),
	}
	return times
}
