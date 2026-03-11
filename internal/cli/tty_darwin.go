//go:build darwin

package cli

import (
	"os"
	"syscall"
	"unsafe"
)

// FlushTTYInput discards unread input from the TTY's kernel buffer, preventing
// stray characters (e.g. repeated Esc presses) from leaking into the shell.
func FlushTTYInput(f *os.File) {
	const TIOCFLUSH = 0x80047410 // macOS ioctl: flush tty (takes *int flags)
	const TCIFLUSH = 1           // Discard received but unread data
	flag := uintptr(TCIFLUSH)
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), TIOCFLUSH, uintptr(unsafe.Pointer(&flag))) //nolint:errcheck
}
