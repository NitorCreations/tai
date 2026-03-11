//go:build unix

package cli

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// OpenTTY opens /dev/tty for dedicated TUI input/output so bubbletea uses a
// dedicated fd rather than sharing with the calling shell's stdin/stdout.
func OpenTTY() (*os.File, error) {
	return os.OpenFile("/dev/tty", os.O_RDWR, 0)
}

// RedirectStderrToNull redirects fd=2 to /dev/null and returns a restore
// function. This suppresses Copilot SDK subprocess noise which would otherwise
// corrupt bubbletea's cursor tracking by writing to the same terminal.
func RedirectStderrToNull() (restore func(), err error) {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		return func() {}, err
	}
	savedStderrFd, dupErr := syscall.Dup(2)
	if dupErr != nil {
		devNull.Close()
		return func() {}, dupErr
	}
	unix.Dup2(int(devNull.Fd()), 2) //nolint:errcheck
	return func() {
		unix.Dup2(savedStderrFd, 2)  //nolint:errcheck
		syscall.Close(savedStderrFd) //nolint:errcheck
		devNull.Close()
	}, nil
}
