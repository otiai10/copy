package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopy_SameFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(src, []byte("important data"), 0600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "hard.txt")
	if err := os.Link(src, link); err != nil {
		t.Skipf("hard links not supported: %v", err)
	}
	sym := filepath.Join(dir, "sym.txt")
	if err := os.Symlink(src, sym); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	for name, dest := range map[string]string{
		"same path":    src,
		"unclean path": filepath.Join(dir, ".", "sub", "..", "source.txt"),
		"hard link":    link,
		"symlink":      sym,
	} {
		if err := Copy(src, dest); err == nil {
			t.Errorf("%s: expected an error, got nil", name)
		}
		got, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "important data" {
			t.Errorf("%s: source changed to %q", name, got)
		}
	}

	other := filepath.Join(dir, "other.txt")
	if err := Copy(src, other); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(other); string(got) != "important data" {
		t.Errorf("normal copy gave %q", got)
	}
}
