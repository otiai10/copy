//go:build darwin

package copyonwrite

import "golang.org/x/sys/unix"

func CopyOnWrite(src, dst string) error {
	return unix.Clonefile(src, dst, 0)
}
