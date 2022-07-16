//go:build windows || js
// +build windows js

package copy

import "os"

func preserveLtimes(srcinfo os.FileInfo, dest string) error {
	//Unsupported
	return nil
}
