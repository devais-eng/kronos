package sync

import "github.com/rotisserie/eris"

var (
	ErrTimeout      = eris.New("timeout")
	ErrNotConnected = eris.New("not connected")
)
