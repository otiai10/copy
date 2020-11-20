package copy

import (
	"fmt"
	"os"
	"syscall"
	"time"
)

func preserveTime(info os.FileInfo, dest string) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}
	mtime := info.ModTime()
	statPtr := info.Sys()
	if statPtr == nil {
		return fmt.Errorf("Error in obtain stat structure")
	}
	stat := statPtr.(*syscall.Stat_t)
	atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
	return os.Chtimes(dest, atime, mtime)
}
