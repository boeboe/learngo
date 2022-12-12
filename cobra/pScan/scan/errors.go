package scan

import "errors"

var (
	ErrExists    = errors.New("Host already in the list")
	ErrNotExists = errors.New("Host not in the list")
)
