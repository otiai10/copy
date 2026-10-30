//go:build !windows && !plan9 && !netbsd && !aix && !illumos && !solaris && !js

package copy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopySymlinkDestinationParents(t *testing.T) {
	for _, target := range []string{"target", "missing", "/absolute/target"} {
		t.Run(target, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "source")
			if err := os.WriteFile(filepath.Join(dir, "target"), []byte("payload"), 0600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, src); err != nil {
				t.Fatal(err)
			}
			for _, missingParents := range []bool{false, true} {
				dest := filepath.Join(dir, "existing-parent-link")
				if missingParents {
					dest = filepath.Join(dir, "new", "nested", "link")
				}
				if err := Copy(src, dest); err != nil {
					t.Errorf("missingParents=%v: Copy: %v", missingParents, err)
					continue
				}
				got, err := os.Readlink(dest)
				if err != nil || got != target {
					t.Errorf("Readlink = %q, %v; want %q, nil", got, err, target)
				}
			}
		})
	}
}

func TestCopyRenamedSymlinkDestinationParents(t *testing.T) {
	for _, workers := range []int64{1, 2} {
		dir := t.TempDir()
		src, dest := filepath.Join(dir, "src"), filepath.Join(dir, "dest")
		if err := os.Mkdir(src, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("target", filepath.Join(src, "link")); err != nil {
			t.Fatal(err)
		}
		err := Copy(src, dest, Options{
			NumOfWorkers: workers,
			RenameDestination: func(src, dest string) (string, error) {
				if filepath.Base(src) == "link" {
					return filepath.Join(filepath.Dir(dest), "new", "nested", "link"), nil
				}
				return dest, nil
			},
		})
		if err != nil {
			t.Errorf("workers=%d: Copy: %v", workers, err)
			continue
		}
		got, err := os.Readlink(filepath.Join(dest, "new", "nested", "link"))
		if err != nil || got != "target" {
			t.Errorf("workers=%d: Readlink = %q, %v; want target, nil", workers, got, err)
		}
	}
}

func TestCopySymlinkParentError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "source")
	if err := os.Symlink("target", src); err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, []byte("keep"), 0600); err != nil {
		t.Fatal(err)
	}
	var reported error
	err := Copy(src, filepath.Join(blocker, "link"), Options{
		OnError: func(src, dest string, err error) error {
			reported = err
			return err
		},
	})
	if err == nil || err != reported {
		t.Fatalf("Copy = %v; callback error = %v", err, reported)
	}
	got, err := os.ReadFile(blocker)
	if err != nil || string(got) != "keep" {
		t.Fatalf("destination parent changed: %q, %v", got, err)
	}
}
