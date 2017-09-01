package copy

import (
	"os"
	"testing"

	. "github.com/otiai10/mint"
)

func TestCopy(t *testing.T) {
	defer os.RemoveAll("testdata/case01")
	err := os.MkdirAll("testdata/case01/bar", os.ModePerm)
	Expect(t, err).ToBe(nil)
	_, err = os.Create("testdata/case01/001.txt")
	Expect(t, err).ToBe(nil)
	_, err = os.Create("testdata/case01/bar/002.txt")
	Expect(t, err).ToBe(nil)
	err = Copy("testdata/case01", "testdata/case01/copy")
	Expect(t, err).ToBe(nil)
	info, err := os.Stat("testdata/case01/copy/bar/002.txt")
	Expect(t, err).ToBe(nil)
	Expect(t, info.IsDir()).ToBe(false)

	When(t, "specified src doesn't exist", func(t *testing.T) {
		err := Copy("not/existing/path", "anywhere")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "specified src is just a file", func(t *testing.T) {
		defer os.RemoveAll("testdata/case01.1")
		os.MkdirAll("testdata/case01.1", os.ModePerm)
		os.Create("testdata/case01.1/001.txt")
		err := Copy("testdata/case01.1/001.txt", "testdata/case01.1/002.txt")
		Expect(t, err).ToBe(nil)
	})
}
