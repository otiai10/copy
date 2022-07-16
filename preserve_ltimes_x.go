//go:build windows || js || plan9
// +build windows js plan9

package copy

import "os"

func preserveLtimes(srcinfo os.FileInfo, dest string) error {
	return nil // Unsupported
}
