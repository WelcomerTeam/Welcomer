package backend

import (
	"errors"
)

var (
	ErrBackendAlreadyExists = errors.New("backend already created")

	ErrReadConfigurationFailure = errors.New("failed to read configuration")
	ErrLoadConfigurationFailure = errors.New("failed to load configuration")

	ErrMissingToken = errors.New("missing token in session")
	ErrMissingUser  = errors.New("missing user in session")

	ErrMissingParameter = errors.New("missing parameter \"%s\" in request")
	ErrWelcomerMissing  = errors.New("bot is missing from server")
	ErrEnsureFailure    = errors.New("failed to ensure guild")
)

// Validation errors.
var (
	ErrRequired                 = errors.New("this field is required")
	ErrChannelInvalid           = errors.New("this channel does not exist")
	ErrInvalidJSON              = errors.New("invalid json")
	ErrInvalidColour            = errors.New("colour format is not recognised")
	ErrInvalidBackground        = errors.New("invalid background")
	ErrInvalidImageAlignment    = errors.New("image alignment is not recognised")
	ErrInvalidImageTheme        = errors.New("image theme is not recognised")
	ErrInvalidProfileBorderType = errors.New("profile border type is not recognised")
)
