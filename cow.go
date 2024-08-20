package copy

import "errors"

// ErrNoCOW is the error returned when copy-on-write is not possible.
//
// To check for this error:
//
//	errors.Is(err, copy.ErrNoCOW)
var ErrNoCOW = errors.New("copy-on-write not possible")
