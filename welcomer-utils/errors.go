package utils

import "errors"

var (
	ErrInvalidColour = errors.New("colour format is not recognised")

	ErrMissingGuild           = errors.New("missing guild")
	ErrMissingChannel         = errors.New("missing channel")
	ErrMissingApplicationUser = errors.New("missing application user")

	ErrInvalidTempChannel = errors.New("channel is not a temporary channel")
)
