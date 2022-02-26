//go:build !go1.16
// +build !go1.16

package copy

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	. "github.com/otiai10/mint"
)

func TestCopy_PathError(t *testing.T) {

	When(t, "too long name is given", func(t *testing.T) {
		dest := "foobar"
		for i := 0; i < 8; i++ {
			dest = dest + dest
		}
		err := Copy("test/data/case00", filepath.Join("test/data/case00", dest))
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err).TypeOf("*os.PathError")
	})

	When(t, "try to create not permitted location", func(t *testing.T) {
		if runtime.GOOS == "windows" || runtime.GOOS == "freebsd" || os.Getenv("TESTCASE") != "" {
			t.Skipf("FIXME: error IS nil here in Windows and FreeBSD")
		}
		err := Copy("test/data/case00", "/case00")
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err).TypeOf("*os.PathError")
	})

	When(t, "try to create a directory on existing file name", func(t *testing.T) {
		err := Copy("test/data/case02", "test/data.copy/case00/README.md")
		Expect(t, err).Not().ToBe(nil)
		Expect(t, err).TypeOf("*os.PathError")
	})
}
