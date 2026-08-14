//go:build !windows && !plan9 && !js

package copy

import (
	"io/fs"

	"golang.org/x/sys/unix"
)

func preserveLtimes(src, dest string, opt Options) error {
	info, err := fs.Lstat(opt.fs(), src)
	if err != nil {
		return err
	}

	if stat, ok := info.Sys().(unix.Stat_t); ok {
		return unix.Lutimes(dest, []unix.Timeval{
			unix.NsecToTimeval(stat.Atim.Nano()),
			unix.NsecToTimeval(stat.Mtim.Nano()),
		})
	}
	return unix.Lutimes(dest, []unix.Timeval{
		unix.NsecToTimeval(0),
		unix.NsecToTimeval(info.ModTime().UnixNano()),
	})
}
