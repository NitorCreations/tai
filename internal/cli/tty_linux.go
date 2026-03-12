//go:build linux

package cli

import (
	"os"
	"syscall"
)

// FlushTTYInput discards unread input from the TTY's kernel buffer, preventing
// stray characters (e.g. repeated Esc presses) from leaking into the shell.
func FlushTTYInput(f *os.File) {
	const TCFLSH = 0x540B                                        // Linux ioctl: flush tty
	const TCIFLUSH = 0                                           // Discard received but unread data
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), TCFLSH, TCIFLUSH) //nolint:errcheck
}
