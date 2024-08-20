//go:build !darwin

package copy

import "os"

const platformSupportsCopyOnWrite = false

// fcopyOnWrite tries to copy a file using the platform's copy-on-write
// mechanism
func fcopyOnWrite(src, dest string, info os.FileInfo, opt Options) (err error) {
	return ErrNoCOW
}
