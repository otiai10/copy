package copy

import (
	"embed"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/otiai10/mint"
)

//go:embed test/data/case18/assets
var assets embed.FS

func TestMain(m *testing.M) {
	setup(m)
	code := m.Run()
	teardown(m)
	os.Exit(code)
}

func teardown(m *testing.M) {
	os.RemoveAll("test/data/case03/case01")
	os.RemoveAll("test/data.copy")
	os.RemoveAll("test/data.copyTime")
	os.RemoveAll("test/owned-by-root") // Do not check the error ;)
	Copy("test/data/case18/assets.backup", "test/data/case18/assets")
	os.RemoveAll("test/data/case18/assets.backup")
}

func TestCopy(t *testing.T) {
	err := Copy("./test/data/case00", "./test/data.copy/case00")
	Expect(t, err).ToBe(nil)
	info, err := os.Stat("./test/data.copy/case00/README.md")
	Expect(t, err).ToBe(nil)
	Expect(t, info).Not().ToBe(nil)
	Expect(t, info.IsDir()).ToBe(false)

	When(t, "specified src doesn't exist", func(t *testing.T) {
		err := Copy("NOT/EXISTING/SOURCE/PATH", "anywhere")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "specified src is just a file", func(t *testing.T) {
		err := Copy("test/data/case01/README.md", "test/data.copy/case01/README.md")
		Expect(t, err).ToBe(nil)
		content, err := ioutil.ReadFile("test/data.copy/case01/README.md")
		Expect(t, err).ToBe(nil)
		Expect(t, string(content)).ToBe("case01 - README.md")
	})

	When(t, "source directory includes symbolic link", func(t *testing.T) {
		err := Copy("test/data/case03", "test/data.copy/case03")
		Expect(t, err).ToBe(nil)
		info, err := os.Lstat("test/data.copy/case03/case01")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(0)
	})

	When(t, "try to copy to an existing path", func(t *testing.T) {
		err := Copy("test/data/case03", "test/data.copy/case03")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to copy READ-not-allowed source", func(t *testing.T) {
		err := Copy("test/data/doesNotExist", "test/data.copy/doesNotExist")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to copy a file to existing path", func(t *testing.T) {
		err := Copy("test/data/case04/README.md", "test/data/case04")
		Expect(t, err).Not().ToBe(nil)
		err = Copy("test/data/case04/README.md", "test/data/case04/README.md/foobar")
		Expect(t, err).Not().ToBe(nil)
	})

	When(t, "try to copy a directory that has no write permission and copy file inside along with it", func(t *testing.T) {
		src := "test/data/case05"
		dest := "test/data.copy/case05"
		err := os.Chmod(src, os.FileMode(0o555))
		Expect(t, err).ToBe(nil)
		err = Copy(src, dest)
		Expect(t, err).ToBe(nil)
		info, err := os.Lstat(dest)
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode().Perm()).ToBe(os.FileMode(0o555))
		err = os.Chmod(dest, 0o755)
		Expect(t, err).ToBe(nil)
	})
	When(t, "file is deleted while copying", func(t *testing.T) {
		src := t.TempDir()
		dest := t.TempDir()

		file := filepath.Join(src, "file")
		f, err := os.Create(file)
		Expect(t, err).ToBe(nil)
		f.Close()

		opt := Options{Skip: func(info os.FileInfo, src, dest string) (bool, error) {
			os.Remove(src)
			return false, nil
		}}
		err = Copy(src, dest, opt)
		Expect(t, err).ToBe(nil)
	})
	When(t, "symlink is deleted while copying", func(t *testing.T) {
		src := t.TempDir()
		dest := t.TempDir()

		Expect(t, os.Symlink(".", filepath.Join(src, "symlink"))).ToBe(nil)

		opt := Options{Skip: func(info os.FileInfo, src, dest string) (bool, error) {
			os.Remove(src)
			return false, nil
		}}
		err = Copy(src, dest, opt)
		Expect(t, err).ToBe(nil)
	})
	When(t, "directory is deleted while copying", func(t *testing.T) {
		src := t.TempDir()
		dest := t.TempDir()

		Expect(t, os.Mkdir(filepath.Join(src, "dir"), 0755)).ToBe(nil)

		opt := Options{Skip: func(info os.FileInfo, src, dest string) (bool, error) {
			os.Remove(src)
			return false, nil
		}}
		err = Copy(src, dest, opt)
		Expect(t, err).ToBe(nil)
	})
}

func TestCopy_NamedPipe(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "js" {
		t.Skip("See https://github.com/otiai10/copy/issues/47")
	}

	When(t, "specified src contains a folder with a named pipe", func(t *testing.T) {
		dest := "test/data.copy/case11"
		err := Copy("test/data/case11", dest)
		Expect(t, err).ToBe(nil)

		info, err := os.Lstat("test/data/case11/foo/bar")
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeNamedPipe != 0).ToBe(true)
		Expect(t, info.Mode().Perm()).ToBe(os.FileMode(0o555))
	})

	When(t, "specified src is a named pipe", func(t *testing.T) {
		dest := "test/data.copy/case11/foo/bar.named"
		err := Copy("test/data/case11/foo/bar", dest)
		Expect(t, err).ToBe(nil)

		info, err := os.Lstat(dest)
		Expect(t, err).ToBe(nil)
		Expect(t, info.Mode()&os.ModeNamedPipe != 0).ToBe(true)
		Expect(t, info.Mode().Perm()).ToBe(os.FileMode(0o555))
	})
}

func TestOptions_OnSymlink(t *testing.T) {
	opt := Options{OnSymlink: func(string) SymlinkAction { return Deep }}
	err := Copy("test/data/case03", "test/data.copy/case03.deep", opt)
	Expect(t, err).ToBe(nil)
	info, err := os.Lstat("test/data.copy/case03.deep/case01")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()&os.ModeSymlink).ToBe(os.FileMode(0))

	opt = Options{OnSymlink: func(string) SymlinkAction { return Shallow }}
	err = Copy("test/data/case03", "test/data.copy/case03.shallow", opt)
	Expect(t, err).ToBe(nil)
	info, err = os.Lstat("test/data.copy/case03.shallow/case01")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(os.FileMode(0))

	opt = Options{OnSymlink: func(string) SymlinkAction { return Skip }}
	err = Copy("test/data/case03", "test/data.copy/case03.skip", opt)
	Expect(t, err).ToBe(nil)
	_, err = os.Stat("test/data.copy/case03.skip/case01")
	Expect(t, os.IsNotExist(err)).ToBe(true)

	err = Copy("test/data/case03", "test/data.copy/case03.default")
	Expect(t, err).ToBe(nil)
	info, err = os.Lstat("test/data.copy/case03.default/case01")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(os.FileMode(0))

	opt = Options{OnSymlink: nil}
	err = Copy("test/data/case03", "test/data.copy/case03.not-specified", opt)
	Expect(t, err).ToBe(nil)
	info, err = os.Lstat("test/data.copy/case03.not-specified/case01")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()&os.ModeSymlink).Not().ToBe(os.FileMode(0))
}

func TestOptions_Skip(t *testing.T) {
	opt := Options{Skip: func(info os.FileInfo, src, dest string) (bool, error) {
		switch {
		case strings.HasSuffix(src, "_skip"):
			return true, nil
		case strings.HasSuffix(src, ".gitfake"):
			return true, nil
		default:
			return false, nil
		}
	}}
	err := Copy("test/data/case06", "test/data.copy/case06", opt)
	Expect(t, err).ToBe(nil)
	info, err := os.Stat("./test/data.copy/case06/dir_skip")
	Expect(t, info).ToBe(nil)
	Expect(t, os.IsNotExist(err)).ToBe(true)

	info, err = os.Stat("./test/data.copy/case06/file_skip")
	Expect(t, info).ToBe(nil)
	Expect(t, os.IsNotExist(err)).ToBe(true)

	info, err = os.Stat("./test/data.copy/case06/README.md")
	Expect(t, info).Not().ToBe(nil)
	Expect(t, err).ToBe(nil)

	info, err = os.Stat("./test/data.copy/case06/repo/.gitfake")
	Expect(t, info).ToBe(nil)
	Expect(t, os.IsNotExist(err)).ToBe(true)

	info, err = os.Stat("./test/data.copy/case06/repo/README.md")
	Expect(t, info).Not().ToBe(nil)
	Expect(t, err).ToBe(nil)

	Because(t, "if Skip func returns error, Copy should be interrupted", func(t *testing.T) {
		errInsideSkipFunc := errors.New("Something wrong inside Skip")
		opt := Options{Skip: func(info os.FileInfo, src, dest string) (bool, error) {
			return false, errInsideSkipFunc
		}}
		err := Copy("test/data/case06", "test/data.copy/case06.01", opt)
		Expect(t, err).ToBe(errInsideSkipFunc)
		files, err := ioutil.ReadDir("./test/data.copy/case06.01")
		Expect(t, err).ToBe(nil)
		Expect(t, len(files)).ToBe(0)
	})
}

func TestOptions_Specials(t *testing.T) {
	if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		err := Copy("/dev/null", "test/data.copy/dev-null", Options{Specials: false})
		Expect(t, err).ToBe(nil)
	}
}

func TestOptions_PermissionControl(t *testing.T) {
	info, err := os.Stat("test/data/case07/dir_0555")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()).ToBe(os.FileMode(0o555) | os.ModeDir)

	info, err = os.Stat("test/data/case07/file_0444")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()).ToBe(os.FileMode(0o444))

	opt := Options{PermissionControl: AddPermission(0o222)}
	err = Copy("test/data/case07", "test/data.copy/case07", opt)
	Expect(t, err).ToBe(nil)

	info, err = os.Stat("test/data.copy/case07/dir_0555")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()).ToBe(os.FileMode(0o555|0o222) | os.ModeDir)

	info, err = os.Stat("test/data.copy/case07/file_0444")
	Expect(t, err).ToBe(nil)
	Expect(t, info.Mode()).ToBe(os.FileMode(0o444 | 0o222))

	When(t, "try to copy a dir owned by other users", func(t *testing.T) {
		err := Copy("test/data/case14", "test/owned-by-root", Options{
			PermissionControl: DoNothing, // ONLY docker tests fail when you comment out this line
		})
		Expect(t, err).ToBe(nil)
	})
}

func TestOptions_Sync(t *testing.T) {
	// With Sync option, each file will be flushed to storage on copying.
	// TODO: Since it's a bit hard to simulate real usecases here. This testcase is nonsense.
	opt := Options{Sync: true}
	err := Copy("test/data/case08", "test/data.copy/case08", opt)
	Expect(t, err).ToBe(nil)
}

func TestOptions_PreserveTimes(t *testing.T) {
	err := Copy("test/data/case09", "test/data.copy/case09")
	Expect(t, err).ToBe(nil)
	opt := Options{PreserveTimes: true}
	err = Copy("test/data/case09", "test/data.copy/case09-preservetimes", opt)
	Expect(t, err).ToBe(nil)

	for _, entry := range []string{"", "README.md", "symlink"} {
		orig, err := os.Stat("test/data/case09/" + entry)
		Expect(t, err).ToBe(nil)
		plain, err := os.Stat("test/data.copy/case09/" + entry)
		Expect(t, err).ToBe(nil)
		preserved, err := os.Stat("test/data.copy/case09-preservetimes/" + entry)
		Expect(t, err).ToBe(nil)
		Expect(t, plain.ModTime().Unix()).Not().ToBe(orig.ModTime().Unix())
		Expect(t, preserved.ModTime().Unix()).ToBe(orig.ModTime().Unix())
	}
}

func TestOptions_OnDirExists(t *testing.T) {
	err := Copy("test/data/case10/dest", "test/data.copy/case10/dest.1")
	Expect(t, err).ToBe(nil)
	err = Copy("test/data/case10/dest", "test/data.copy/case10/dest.2")
	Expect(t, err).ToBe(nil)
	err = Copy("test/data/case10/dest", "test/data.copy/case10/dest.3")
	Expect(t, err).ToBe(nil)

	opt := Options{}

	opt.OnDirExists = func(src, dest string) DirExistsAction {
		return Merge
	}
	err = Copy("test/data/case10/src", "test/data.copy/case10/dest.1", opt)
	Expect(t, err).ToBe(nil)
	err = Copy("test/data/case10/src", "test/data.copy/case10/dest.1", opt)
	Expect(t, err).ToBe(nil)
	b, err := ioutil.ReadFile("test/data.copy/case10/dest.1/" + "foo/" + "text_aaa")
	Expect(t, err).ToBe(nil)
	Expect(t, string(b)).ToBe("This is text_aaa from src")
	stat, err := os.Stat("test/data.copy/case10/dest.1/foo/text_eee")
	Expect(t, err).ToBe(nil)
	Expect(t, stat).Not().ToBe(nil)

	opt.OnDirExists = func(src, dest string) DirExistsAction {
		return Replace
	}
	err = Copy("test/data/case10/src", "test/data.copy/case10/dest.2", opt)
	Expect(t, err).ToBe(nil)
	err = Copy("test/data/case10/src", "test/data.copy/case10/dest.2", opt)
	Expect(t, err).ToBe(nil)
	b, err = ioutil.ReadFile("test/data.copy/case10/dest.2/" + "foo/" + "text_aaa")
	Expect(t, err).ToBe(nil)
	Expect(t, string(b)).ToBe("This is text_aaa from src")
	stat, err = os.Stat("test/data.copy/case10/dest.2/foo/text_eee")
	Expect(t, os.IsNotExist(err)).ToBe(true)
	Expect(t, stat).ToBe(nil)

	opt.OnDirExists = func(src, dest string) DirExistsAction {
		return Untouchable
	}
	err = Copy("test/data/case10/src", "test/data.copy/case10/dest.3", opt)
	Expect(t, err).ToBe(nil)
	b, err = ioutil.ReadFile("test/data.copy/case10/dest.3/" + "foo/" + "text_aaa")
	Expect(t, err).ToBe(nil)
	Expect(t, string(b)).ToBe("This is text_aaa from dest")

	When(t, "PreserveTimes is true with Untouchable", func(t *testing.T) {
		opt := Options{
			OnDirExists:   func(src, dest string) DirExistsAction { return Untouchable },
			PreserveTimes: true,
		}
		err = Copy("test/data/case10/src", "test/data.copy/case10/dest.3", opt)
		Expect(t, err).ToBe(nil)
	})
}

func TestOptions_CopyBufferSize(t *testing.T) {
	opt := Options{
		CopyBufferSize: 512,
	}

	err := Copy("test/data/case12", "test/data.copy/case12", opt)
	Expect(t, err).ToBe(nil)

	content, err := ioutil.ReadFile("test/data.copy/case12/README.md")
	Expect(t, err).ToBe(nil)
	Expect(t, string(content)).ToBe("case12 - README.md")
}

func TestOptions_PreserveOwner(t *testing.T) {
	opt := Options{PreserveOwner: true}
	err := Copy("test/data/case13", "test/data.copy/case13", opt)
	Expect(t, err).ToBe(nil)
}

func TestOptions_CopyRateLimit(t *testing.T) {

	file, err := os.Create("test/data/case16/large.file")
	if err != nil {
		t.Errorf("failed to create test file: %v", err)
		return
	}

	size := int64(100 * 1024) // 100 KB
	if err := file.Truncate(size); err != nil {
		t.Errorf("failed to truncate test file: %v", err)
		t.SkipNow()
		return
	}

	opt := Options{WrapReader: func(src io.Reader) io.Reader {
		return &SleepyReader{src, 1}
	}}

	start := time.Now()
	err = Copy("test/data/case16", "test/data.copy/case16", opt)
	elapsed := time.Since(start)
	Expect(t, err).ToBe(nil)
	Expect(t, elapsed > 5*time.Second).ToBe(true)
}

func TestOptions_OnFileError(t *testing.T) {
	opt := Options{
		OnError: nil,
	}

	// existing, process nromally
	err := Copy("test/data/case17", "test/data.copy/case17", opt)
	Expect(t, err).ToBe(nil)

	// not existing, process err
	err = Copy("test/data/case17/non-existing", "test/data.copy/case17/non-existing", opt)
	Expect(t, os.IsNotExist(err)).ToBe(true)

	_, err = os.Stat("test/data.copy/case17/non-existing")
	Expect(t, os.IsNotExist(err)).ToBe(true)

	// existing, nil err not passed
	opt.OnError = func(_, _ string, err error) error {
		return err
	}
	err = Copy("test/data/case17", "test/data.copy/case17", opt)
	Expect(t, err).ToBe(nil)

	// not existing, process err
	opt.OnError = func(_, _ string, err error) error { return err }
	err = Copy("test/data/case17/non-existing", "test/data.copy/case17/non-existing", opt)
	Expect(t, os.IsNotExist(err)).ToBe(true)

	_, err = os.Stat("test/data.copy/case17/non-existing")
	Expect(t, os.IsNotExist(err)).ToBe(true)

	// not existing, ignore err
	opt.OnError = func(_, _ string, err error) error { return nil }
	err = Copy("test/data/case17/non-existing", "test/data.copy/case17/non-existing", opt)
	Expect(t, err).ToBe(nil)

	_, err = os.Stat("test/data.copy/case17/non-existing")
	Expect(t, os.IsNotExist(err)).ToBe(true)
}

func TestOptions_FS(t *testing.T) {
	os.RemoveAll("test/data/case18/assets")
	err := Copy("test/data/case18/assets", "test/data.copy/case18/assets", Options{
		FS:                assets,
		PermissionControl: AddPermission(200), // FIXME
	})
	Expect(t, err).ToBe(nil)
}

type SleepyReader struct {
	src io.Reader
	sec time.Duration
}

func (r *SleepyReader) Read(p []byte) (int, error) {
	n, e := r.src.Read(p)
	if e != nil && e != io.EOF {
		return n, e
	}
	if n > 0 {
		time.Sleep(time.Second * r.sec)
	}
	return n, e
}

func TestOptions_NumOfWorkers(t *testing.T) {
	opt := Options{NumOfWorkers: 3}
	err := Copy("test/data/case19", "test/data.copy/case19", opt)
	Expect(t, err).ToBe(nil)
}

func TestOptions_PreferConcurrent(t *testing.T) {
	opt := Options{NumOfWorkers: 4, PreferConcurrent: func(sd, dd string) (bool, error) {
		return strings.HasSuffix(sd, "concurrent"), nil
	}}
	err := Copy("test/data/case19", "test/data.copy/case19_preferconcurrent", opt)
	Expect(t, err).ToBe(nil)
}
