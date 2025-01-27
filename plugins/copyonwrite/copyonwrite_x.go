//go:build !darwin

package copyonwrite

import "errors"

func CopyOnWrite(src, dst string) error {
	return errors.New("copy-on-write is not supported on this platform")
}
