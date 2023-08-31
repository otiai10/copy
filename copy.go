package copy

import (
	"go.uber.org/multierr"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type timespec struct {
	Mtime time.Time
	Atime time.Time
	Ctime time.Time
}

// Copy copies src to dest, doesn't matter if src is a directory or a file.
func Copy(src, dest string, opts ...Options) error {
	opt := assureOptions(src, dest, opts...)

	var numConcurrentCopies uint = 1
	if opt.Concurrency > 1 {
		numConcurrentCopies = opt.Concurrency
	}

	inCh := make(chan workerInput)
	outCh := make(chan workerOutput)
	errCh := make(chan error)
	go startWorkers(numConcurrentCopies, inCh, outCh)
	go processResults(outCh, errCh)

	if opt.FS != nil {
		info, err := fs.Stat(opt.FS, src)
		if err != nil {
			return onError(src, dest, err, opt)
		}
		return switchboard(src, dest, info, opt, inCh)
	}
	info, err := os.Lstat(src)
	if err != nil {
		return onError(src, dest, err, opt)
	}

	err = switchboard(src, dest, info, opt, inCh)
	if err != nil {
		close(inCh)
		close(outCh)
		return err
	}
	close(inCh)

	return <-errCh
}

// switchboard switches proper copy functions regarding file type, etc...
// If there would be anything else here, add a case to this switchboard.
func switchboard(src, dest string, info os.FileInfo, opt Options, inCh chan workerInput) (err error) {
	if info.Mode()&os.ModeDevice != 0 && !opt.Specials {
		return onError(src, dest, err, opt)
	}

	switch {
	case info.Mode()&os.ModeSymlink != 0:
		err = onsymlink(src, dest, opt, inCh)
	case info.IsDir():
		err = dcopy(src, dest, info, opt, inCh)
	case info.Mode()&os.ModeNamedPipe != 0:
		err = pcopy(dest, info)
	default:
		inCh <- workerInput{src, dest, info, opt}
	}

	return onError(src, dest, err, opt)
}

// copyNextOrSkip decide if this src should be copied or not.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func copyNextOrSkip(src, dest string, info os.FileInfo, opt Options, inCh chan workerInput) error {
	if opt.Skip != nil {
		skip, err := opt.Skip(info, src, dest)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
	}
	return switchboard(src, dest, info, opt, inCh)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo, opt Options) (err error) {

	var readcloser io.ReadCloser
	if opt.FS != nil {
		readcloser, err = opt.FS.Open(src)
	} else {
		readcloser, err = os.Open(src)
	}
	if err != nil {
		if os.IsNotExist(err) {
			return nil
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
		return err
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
		return err
	}

	if opt.Sync {
		err = f.Sync()
	}

	if opt.PreserveOwner {
		if err := preserveOwner(src, dest, info); err != nil {
			return err
		}
	}
	if opt.PreserveTimes {
		if err := preserveTimes(info, dest); err != nil {
			return err
		}
	}

	return
}

// dcopy is for a directory,
// with scanning contents inside the directory
// and pass everything to "copy" recursively.
func dcopy(srcdir, destdir string, info os.FileInfo, opt Options, inCh chan workerInput) (err error) {
	if skip, err := onDirExists(opt, srcdir, destdir); err != nil {
		return err
	} else if skip {
		return nil
	}

	// Make dest dir with 0755 so that everything writable.
	chmodfunc, err := opt.PermissionControl(info, destdir)
	if err != nil {
		return err
	}
	defer chmodfunc(&err)

	var contents []os.FileInfo
	if opt.FS != nil {
		entries, err := fs.ReadDir(opt.FS, srcdir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				return err
			}
			contents = append(contents, info)
		}
	} else {
		contents, err = ioutil.ReadDir(srcdir)
	}

	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())

		if err = copyNextOrSkip(cs, cd, content, opt, inCh); err != nil {
			// If any error, exit immediately
			return
		}
	}

	if opt.PreserveTimes {
		if err := preserveTimes(info, destdir); err != nil {
			return err
		}
	}

	if opt.PreserveOwner {
		if err := preserveOwner(srcdir, destdir, info); err != nil {
			return err
		}
	}

	return
}

func onDirExists(opt Options, srcdir, destdir string) (bool, error) {
	_, err := os.Stat(destdir)
	if err == nil && opt.OnDirExists != nil && destdir != opt.intent.dest {
		switch opt.OnDirExists(srcdir, destdir) {
		case Replace:
			if err := os.RemoveAll(destdir); err != nil {
				return false, err
			}
		case Untouchable:
			return true, nil
		} // case "Merge" is default behaviour. Go through.
	} else if err != nil && !os.IsNotExist(err) {
		return true, err // Unwelcome error type...!
	}
	return false, nil
}

func onsymlink(src, dest string, opt Options, inCh chan workerInput) error {
	switch opt.OnSymlink(src) {
	case Shallow:
		if err := lcopy(src, dest); err != nil {
			return err
		}
		if opt.PreserveTimes {
			return preserveLtimes(src, dest)
		}
		return nil
	case Deep:
		orig, err := os.Readlink(src)
		if err != nil {
			return err
		}
		info, err := os.Lstat(orig)
		if err != nil {
			return err
		}
		return copyNextOrSkip(orig, dest, info, opt, inCh)
	case Skip:
		fallthrough
	default:
		return nil // do nothing
	}
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string) error {
	src, err := os.Readlink(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Symlink(src, dest)
}

// fclose ANYHOW closes file,
// with asiging error raised during Close,
// BUT respecting the error already reported.
func fclose(f io.Closer, reported *error) {
	if err := f.Close(); *reported == nil {
		*reported = err
	}
}

// onError lets caller to handle errors
// occured when copying a file.
func onError(src, dest string, err error, opt Options) error {
	if opt.OnError == nil {
		return err
	}

	return opt.OnError(src, dest, err)
}

type workerInput struct {
	src  string
	dest string
	info os.FileInfo
	opt  Options
}

type workerOutput error

func startWorkers(numWorkers uint, inCh chan workerInput, outCh chan workerOutput) {
	var wg sync.WaitGroup
	for workerID := uint(0); workerID < numWorkers; workerID++ {
		wg.Add(1)
		go worker(&wg, inCh, outCh)
	}
	wg.Wait()
	close(outCh)
}

func worker(wg *sync.WaitGroup, inCh chan workerInput, outCh chan workerOutput) {
	for i := range inCh {
		outCh <- fcopy(i.src, i.dest, i.info, i.opt)
	}
	wg.Done()
}

func processResults(out chan workerOutput, result chan error) {
	var err error
	for o := range out {
		err = multierr.Append(err, o)
	}
	result <- err
}
