package welcomer

import "errors"

var (
	ErrInvalidColour = errors.New("colour format is not recognised")

	ErrMissingGuild           = errors.New("missing guild")
	ErrMissingChannel         = errors.New("missing channel")
	ErrMissingUser            = errors.New("missing user")
	ErrMissingApplicationUser = errors.New("missing application user")

	ErrInvalidURL = errors.New("invalid url")

	ErrInvalidTempChannel = errors.New("channel is not a temporary channel")
)
