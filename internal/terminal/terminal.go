package terminal

import (
	"fmt"
	"os"
)

// Terminal manages OS-level I/O, raw mode, and screen clearing
type Terminal struct {
	// Add state for original terminal settings to restore
}

// New initializes a new Terminal in raw mode
func New() (*Terminal, error) {
	t := &Terminal{}
	if err := t.EnableRawMode(); err != nil {
		return nil, err
	}
	return t, nil
}

// EnableRawMode enables raw input mode on the terminal
func (t *Terminal) EnableRawMode() error {
	// TODO: use termios/sys-level logic to set raw mode
	return nil
}

// Restore restores the terminal to its original state
func (t *Terminal) Restore() {
	// TODO: restore original settings
}

// Clear clears the screen and repositions the cursor
func (t *Terminal) Clear() {
	fmt.Fprint(os.Stdout, "\x1b[2J\x1b[H")
}
