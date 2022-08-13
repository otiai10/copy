//go:build !windows && !plan9 && !js
// +build !windows,!plan9,!js

package copy

import (
	"os"
	"testing"

	. "github.com/otiai10/mint"
	"golang.org/x/sys/unix"
)

func TestOptions_PreserveLTimes(t *testing.T) {
	err := Copy("test/data/case15", "test/data.copy/case15")
	Expect(t, err).ToBe(nil)
	opt := Options{PreserveTimes: true}
	err = Copy("test/data/case15", "test/data.copy/case15-preserveltimes", opt)
	Expect(t, err).ToBe(nil)

	orig, err := os.Lstat("test/data/case15/symlink")
	Expect(t, err).ToBe(nil)
	plain, err := os.Lstat("test/data.copy/case15/symlink")
	Expect(t, err).ToBe(nil)
	preserved, err := os.Lstat("test/data.copy/case15-preserveltimes/symlink")
	Expect(t, err).ToBe(nil)
	Expect(t, plain.ModTime().Unix()).Not().ToBe(orig.ModTime().Unix())
	Expect(t, preserved.ModTime().Unix()).ToBe(orig.ModTime().Unix())
}

func TestOptions_PreserveLTimesErrorReturn(t *testing.T) {
	err := preserveLtimes("doesnotexist_original.txt", "doesnotexist_copy.txt")
	Expect(t, err).ToBe(unix.ENOENT)
	Expect(t, os.IsNotExist(err)).ToBe(true)
}
