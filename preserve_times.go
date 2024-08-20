package copy

import (
	"os"
	"time"
)

func preserveTimes(srcinfo os.FileInfo, dest string) error {
	spec := getTimeSpec(srcinfo)
	if err := os.Chtimes(dest, spec.Atime, spec.Mtime); err != nil {
		return err
	}
	return nil
}

func touchTimes(dest string) error {
	now := time.Now()
	if err := os.Chtimes(dest, now, now); err != nil {
		return err
	}
	return nil
}
