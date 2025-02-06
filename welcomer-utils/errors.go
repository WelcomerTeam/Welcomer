package utils

import "errors"

var (
	ErrInvalidJSON       = errors.New("invalid json")
	ErrInvalidColour     = errors.New("colour format is not recognised")
	ErrInvalidBackground = errors.New("invalid background")

	ErrMissingGuild           = errors.New("missing guild")
	ErrMissingChannel         = errors.New("missing channel")
	ErrMissingApplicationUser = errors.New("missing application user")

	ErrInvalidTempChannel = errors.New("channel is not a temporary channel")
)
