// +build !windows

package copy

import (
	"os"
	"testing"
	"syscall"

	. "github.com/otiai10/mint"
)

func setup(m *testing.M) {
	os.MkdirAll("test/data.copy", os.ModePerm)
	os.Symlink("test/data/case01", "test/data/case03/case01")
	os.Chmod("test/data/case07/dir_0555", 0555)
	os.Chmod("test/data/case07/file_0444", 0444)
	syscall.Mkfifo("test/data/case11/foo/bar", 0555)
}

func testPipes(t *testing.T) {
	When(t, "specified src contains a folder with a named pipe", func(t *testing.T) {
		dest:= "test/data.copy/case11"
		err := Copy("test/data/case11", dest)
		Expect(t, err).ToBe(nil)

		info, err := os.Lstat("test/data/case11/foo/bar")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeNamedPipe != 0).ToBe(true)
		Expect(t, info.Mode().Perm()).ToBe(os.FileMode(0555))
	})

	When(t, "specified src is a named pipe", func(t *testing.T) {
		dest:= "test/data.copy/case11/foo/bar.named"
		err := Copy("test/data/case11/foo/bar", dest)
		Expect(t, err).ToBe(nil)

		info, err := os.Lstat(dest)
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeNamedPipe != 0).ToBe(true)
		Expect(t, info.Mode().Perm()).ToBe(os.FileMode(0555))
	})
}