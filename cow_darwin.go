//go:build darwin

package copy

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

const platformSupportsCopyOnWrite = true

// Darwin implementation uses the `clonefile` syscall:
// https://www.manpagez.com/man/2/clonefile/
//
// Considerations:
//  * Ownership is not preserved.
//  * Setuid and Setgid are not preserved.
//  * Flag CLONE_NOFOLLOW is not used, we use lcopy instead of fcopy for
//    symbolic links.

func fcopyOnWrite(src, dest string, info os.FileInfo, opt Options) (err error) {
	err = unix.Clonefile(src, dest, 0)

	// If the error is the file already exists, delete it and try again.
	// When/if an option for OnFileExists is added, this can be handled
	// differently.
	if errors.Is(err, os.ErrExist) {
		if err = os.Remove(dest); err != nil {
			return fmt.Errorf("%w: %w", ErrNoCOW, err)
		}

		err = unix.Clonefile(src, dest, 0) // retry
	}

	// Return an ErrNoCOW if copy-on-write is not possible.
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNoCOW, err)
	}

	// Copy-on-write preserves the modtime by default.
	// If PreserveTimes is not true, update the time to now.
	if !opt.PreserveTimes {
		if err = touchTimes(dest); err != nil {
			return err
		}
	}

	// Apply permission control to the copied file.
	//
	// Unlike fcopyBytes, this is done after-the-fact since we don't
	// actually open the dest file at any point.
	err = applyPermissionControl(dest, info, opt)

	return
}
