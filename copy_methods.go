package copy

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

// ErrUnsupportedCopyMethod is returned when the FileCopyMethod specified in
// Options is not supported.
var ErrUnsupportedCopyMethod = errors.New(
	"copy method not supported",
)

// CopyBytes copies the file contents by reading the source file into a buffer,
// then writing the buffer back to the destination file.
var CopyBytes = FileCopyMethod{
	fcopy: func(src, dest string, info os.FileInfo, opt Options) (err error, skipFile bool) {

		var readcloser io.ReadCloser
		if opt.FS != nil {
			readcloser, err = opt.FS.Open(src)
		} else {
			readcloser, err = os.Open(src)
		}
		if err != nil {
			if os.IsNotExist(err) {
				return nil, true
			}
			return
		}
		defer fclose(readcloser, &err)

		if err = os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			return
		}

		f, err := os.Create(dest)
		if err != nil {
			return
		}
		defer fclose(f, &err)

		chmodfunc, err := opt.PermissionControl(info, dest)
		if err != nil {
			return err, false
		}
		chmodfunc(&err)

		var buf []byte = nil
		var w io.Writer = f
		var r io.Reader = readcloser

		if opt.WrapReader != nil {
			r = opt.WrapReader(r)
		}

		if opt.CopyBufferSize != 0 {
			buf = make([]byte, opt.CopyBufferSize)
			// Disable using `ReadFrom` by io.CopyBuffer.
			// See https://github.com/otiai10/copy/pull/60#discussion_r627320811 for more details.
			w = struct{ io.Writer }{f}
			// r = struct{ io.Reader }{s}
		}

		if _, err = io.CopyBuffer(w, r, buf); err != nil {
			return err, false
		}

		if opt.Sync {
			err = f.Sync()
		}

		if opt.PreserveOwner {
			if err := preserveOwner(src, dest, info); err != nil {
				return err, false
			}
		}
		if opt.PreserveTimes {
			if err := preserveTimes(info, dest); err != nil {
				return err, false
			}
		}

		return
	},
}
