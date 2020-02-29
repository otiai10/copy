package copy

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	// tmpPermissionForDirectory makes the destination directory writable,
	// so that stuff can be copied recursively even if any original directory is NOT writable.
	// See https://github.com/otiai10/copy/pull/9 for more information.
	tmpPermissionForDirectory = os.FileMode(0755)
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string) error {
	return CopyButSkipSome(src, dest, nil)
}

// CopyButSkipSome copies src to dest and skipping toSkip, doesn't matter if src is a directory or a file
func CopyButSkipSome(src, dest string, toSkip []string) error {
	toSkipMap := make(map[string]struct{})
	for i := 0; i < len(toSkip); i++ {
		toSkipMap[toSkip[i]] = struct{}{}
	}

	fmt.Println(toSkipMap)

	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	return copy(src, dest, toSkipMap, info)
}

// copy dispatches copy-funcs according to the mode.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func copy(src, dest string, toSkip map[string]struct{}, info os.FileInfo) error {
	_, ock := toSkip[src]
	fmt.Println(src)
	fmt.Println(ock)

	if _, isToSkip := toSkip[src]; isToSkip {
		return nil
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return lcopy(src, dest, info)
	}

	if info.IsDir() {
		return dcopy(src, dest, toSkip, info)
	}
	return fcopy(src, dest, info)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo) (err error) {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer fclose(f, &err)

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer fclose(s, &err)

	_, err = io.Copy(f, s)
	return err
}

// dcopy is for a directory,
// with scanning contents inside the directory
// and pass everything to "copy" recursively.
func dcopy(srcdir, destdir string, toSkip map[string]struct{}, info os.FileInfo) (err error) {

	originalMode := info.Mode()

	// Make dest dir with 0755 so that everything writable.
	if err := os.MkdirAll(destdir, tmpPermissionForDirectory); err != nil {
		return err
	}
	// Recover dir mode with original one.
	defer chmod(destdir, originalMode, &err)

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := copy(cs, cd, toSkip, content); err != nil {
			// If any error, exit immediately
			return err
		}
	}

	return nil
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string, info os.FileInfo) error {
	src, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}

// fclose ANYHOW closes file,
// with asiging error occured BUT respecting the error already reported.
func fclose(f *os.File, reported *error) {
	if err := f.Close(); *reported == nil {
		*reported = err
	}
}

// chmod ANYHOW changes file mode,
// with asiging error occured BUT respecting the error already reported.
func chmod(dir string, mode os.FileMode, reported *error) {
	if err := os.Chmod(dir, mode); *reported == nil {
		*reported = err
	}
}
