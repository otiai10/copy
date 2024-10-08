package copy

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type timespec struct {
	Mtime time.Time
	Atime time.Time
	Ctime time.Time
}

// Copy copies src to dest, doesn't matter if src is a directory or a file.
func Copy(src, dest string, opts ...Options) error {
	opt := assureOptions(src, dest, opts...)
	if opt.NumOfWorkers > 1 {
		opt.intent.sem = semaphore.NewWeighted(opt.NumOfWorkers)
		opt.intent.ctx = context.Background()
	}
	if opt.FS != nil {
		info, err := fs.Stat(opt.FS, src)
		if err != nil {
			return onError(src, dest, err, opt)
		}
		return switchboard(src, dest, info, opt)
	}
	info, err := os.Lstat(src)
	if err != nil {
		return onError(src, dest, err, opt)
	}
	return switchboard(src, dest, info, opt)
}

// switchboard switches proper copy functions regarding file type, etc...
// If there would be anything else here, add a case to this switchboard.
func switchboard(src, dest string, info os.FileInfo, opt Options) (err error) {
	if info.Mode()&os.ModeDevice != 0 && !opt.Specials {
		return onError(src, dest, err, opt)
	}

	if opt.RenameDestination != nil {
		if dest, err = opt.RenameDestination(src, dest); err != nil {
			return onError(src, dest, err, opt)
		}
	}

	switch {
	case info.Mode()&os.ModeSymlink != 0:
		err = onsymlink(src, dest, opt)
	case info.IsDir():
		err = dcopy(src, dest, info, opt)
	case info.Mode()&os.ModeNamedPipe != 0:
		err = pcopy(dest, info)
	default:
		err = fcopy(src, dest, info, opt)
	}

	return onError(src, dest, err, opt)
}

// copyNextOrSkip decide if this src should be copied or not.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func copyNextOrSkip(src, dest string, info os.FileInfo, opt Options) error {
	if opt.Skip != nil {
		skip, err := opt.Skip(info, src, dest)
		if err != nil {
			return err
		}
		if skip {
			return nil
		}
	}
	return switchboard(src, dest, info, opt)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo, opt Options) (err error) {
	if err = os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return
	}

	// Use FileCopyMethod to do copy.
	err, skipFile := opt.FileCopyMethod.fcopy(src, dest, info, opt)
	if skipFile {
		return nil
	}

	if err != nil {
		return err
	}

	// Change file permissions.
	chmodfunc, err := opt.PermissionControl(info, dest)
	if err != nil {
		return err
	}

	chmodfunc(&err)
	if err != nil {
		return err
	}

	// Preserve file ownership and times.
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
func dcopy(srcdir, destdir string, info os.FileInfo, opt Options) (err error) {
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

	var entries []fs.DirEntry
	if opt.FS != nil {
		entries, err = fs.ReadDir(opt.FS, srcdir)
		if err != nil {
			return err
		}
	} else {
		entries, err = os.ReadDir(srcdir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
	}

	contents := make([]fs.FileInfo, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			return err
		}
		contents = append(contents, info)
	}

	if yes, err := shouldCopyDirectoryConcurrent(opt, srcdir, destdir); err != nil {
		return err
	} else if yes {
		if err := dcopyConcurrent(srcdir, destdir, contents, opt); err != nil {
			return err
		}
	} else {
		if err := dcopySequential(srcdir, destdir, contents, opt); err != nil {
			return err
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

func dcopySequential(srcdir, destdir string, contents []os.FileInfo, opt Options) error {
	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())

		if err := copyNextOrSkip(cs, cd, content, opt); err != nil {
			// If any error, exit immediately
			return err
		}
	}
	return nil
}

// Copy this directory concurrently regarding semaphore of opt.intent
func dcopyConcurrent(srcdir, destdir string, contents []os.FileInfo, opt Options) error {
	group, ctx := errgroup.WithContext(opt.intent.ctx)
	getRoutine := func(cs, cd string, content os.FileInfo) func() error {
		return func() error {
			if content.IsDir() {
				return copyNextOrSkip(cs, cd, content, opt)
			}
			if err := opt.intent.sem.Acquire(ctx, 1); err != nil {
				return err
			}
			err := copyNextOrSkip(cs, cd, content, opt)
			opt.intent.sem.Release(1)
			return err
		}
	}
	for _, content := range contents {
		csd := filepath.Join(srcdir, content.Name())
		cdd := filepath.Join(destdir, content.Name())
		group.Go(getRoutine(csd, cdd, content))
	}
	return group.Wait()
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

func onsymlink(src, dest string, opt Options) error {
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
		if !filepath.IsAbs(orig) {
			// orig is a relative link: need to add src dir to orig
			orig = filepath.Join(filepath.Dir(src), orig)
		}
		info, err := os.Lstat(orig)
		if err != nil {
			return err
		}
		return copyNextOrSkip(orig, dest, info, opt)
	case Skip:
		fallthrough
	default:
		return nil // do nothing
	}
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string) error {
	orig, err := os.Readlink(src)
	// @See https://github.com/otiai10/copy/issues/111
	// TODO: This might be controlled by Options in the future.
	if err != nil {
		if os.IsNotExist(err) { // Copy symlink even if not existing
			return os.Symlink(src, dest)
		}
		return err
	}

	// @See https://github.com/otiai10/copy/issues/132
	// TODO: Control by SymlinkExistsAction
	if _, err := os.Lstat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return err
		}
	}

	return os.Symlink(orig, dest)
}

// fclose ANYHOW closes file,
// with assigning error raised during Close,
// BUT respecting the error already reported.
func fclose(f io.Closer, reported *error) {
	if err := f.Close(); *reported == nil {
		*reported = err
	}
}

// onError lets caller to handle errors
// occurred when copying a file.
func onError(src, dest string, err error, opt Options) error {
	if opt.OnError == nil {
		return err
	}

	return opt.OnError(src, dest, err)
}
