package utils

import "fmt"

var (
	ErrInvalidJSON       = fmt.Errorf("invalid json")
	ErrInvalidColour     = fmt.Errorf("colour format is not recognised")
	ErrInvalidBackground = fmt.Errorf("invalid background")

	ErrMissingGuild           = fmt.Errorf("missing guild")
	ErrMissingChannel         = fmt.Errorf("missing channel")
	ErrMissingApplicationUser = fmt.Errorf("missing application user")

	ErrInvalidTempChannel = fmt.Errorf("channel is not a temporary channel")
)
