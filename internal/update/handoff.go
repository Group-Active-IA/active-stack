package update

import "errors"

// ErrHandoff tells the CLI or TUI to exit so a verified Windows helper can
// replace the executable after its process handle is released.
var ErrHandoff = errors.New("update handed off to helper")
