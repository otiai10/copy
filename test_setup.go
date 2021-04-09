// +build !windows

package copy

import (
	"log"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func setup(m *testing.M) {
	os.MkdirAll("test/data.copy", os.ModePerm)
	os.Symlink("../case01", "test/data/case03/relative")
	os.Symlink("../case03", "test/data/case03/relative")
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get cwd: %v", err)
	}
	os.Symlink(filepath.Join(cwd, "test/data/case01"), "test/data/case03/absolute")
	os.Chmod("test/data/case07/dir_0555", 0555)
	os.Chmod("test/data/case07/file_0444", 0444)
	syscall.Mkfifo("test/data/case11/foo/bar", 0555)
}
