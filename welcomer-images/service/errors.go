package service

import "fmt"

var (
	ErrMissingFrames     = fmt.Errorf("no frames to encode")
	ErrNoFontFound       = fmt.Errorf("no font found")
	ErrNotImplemented    = fmt.Errorf("not yet implemented")
	ErrAvatarFetchFailed = fmt.Errorf("failed to fetch avatar resource")
	ErrInvalidURL        = fmt.Errorf("url is invalid or untrusted")

	ErrInvalidHorizontalAlignment = fmt.Errorf("unknown horizontal alignment")
	ErrInvalidVerticalAlignment   = fmt.Errorf("unknown vertical alignment")
	ErrUnknownProfileBorderType   = fmt.Errorf("unknown profile border type")
	ErrUnknownProfileFloat        = fmt.Errorf("unknown profile float")
)
