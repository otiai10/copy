package copy

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return dcopy(src, dest, info)
	}
	return fcopy(src, dest, info)
}

func fcopy(src, dest string, info os.FileInfo) error {

	if info == nil {
		i, err := os.Stat(src)
		if err != nil {
			return err
		}
		info = i
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	if _, err = io.Copy(f, s); err != nil {
		return err
	}
	return nil
}

func dcopy(src, dest string, info os.FileInfo) error {
	if info == nil {
		i, err := os.Stat(src)
		if err != nil {
			return err
		}
		info = i
	}
	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}
	infos, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	for _, info := range infos {
		if info.IsDir() {
			dcopy(
				filepath.Join(src, info.Name()),
				filepath.Join(dest, info.Name()),
				info,
			)
		} else {
			fcopy(
				filepath.Join(src, info.Name()),
				filepath.Join(dest, info.Name()),
				info,
			)
		}
	}
	return nil
}
