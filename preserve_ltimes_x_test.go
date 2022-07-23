//go:build windows || js || plan9
// +build windows js plan9

package copy

import (
	"os"
	"testing"

	. "github.com/otiai10/mint"
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
	Expect(t, preserved.ModTime().Unix()).Not().ToBe(orig.ModTime().Unix())
}
