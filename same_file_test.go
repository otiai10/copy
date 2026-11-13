package copy

import (
	"os"
	"path/filepath"
	"testing"
)

const sameFileData = "important data"

func setupSameFile(t *testing.T) (dir, src string) {
	t.Helper()
	dir = t.TempDir()
	src = filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte(sameFileData), 0600); err != nil {
		t.Fatal(err)
	}
	return dir, src
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Errorf("%s: got %q, want %q", path, got, want)
	}
}

func TestCopy_SameFile(t *testing.T) {
	for name, alias := range map[string]func(t *testing.T, dir, src string) string{
		"same path": func(t *testing.T, dir, src string) string { return src },
		"unclean path": func(t *testing.T, dir, src string) string {
			return filepath.Join(dir, ".", "sub", "..", "source.txt")
		},
		"hard link": func(t *testing.T, dir, src string) string {
			link := filepath.Join(dir, "hard.txt")
			if err := os.Link(src, link); err != nil {
				t.Skipf("hard links not supported: %v", err)
			}
			return link
		},
		"symlink": func(t *testing.T, dir, src string) string {
			sym := filepath.Join(dir, "sym.txt")
			if err := os.Symlink(src, sym); err != nil {
				t.Skipf("symlinks not supported: %v", err)
			}
			return sym
		},
	} {
		t.Run(name, func(t *testing.T) {
			dir, src := setupSameFile(t)
			dest := alias(t, dir, src)
			if err := Copy(src, dest); err == nil {
				t.Error("expected an error, got nil")
			}
			assertContent(t, src, sameFileData)
		})
	}
}

// Options.FS backed by os.DirFS can reach the same file as dest.
func TestCopy_SameFile_DirFS(t *testing.T) {
	dir, src := setupSameFile(t)
	if err := Copy("source.txt", src, Options{FS: os.DirFS(dir)}); err == nil {
		t.Error("expected an error, got nil")
	}
	assertContent(t, src, sameFileData)
}
