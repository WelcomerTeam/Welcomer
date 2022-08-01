package backend

import "errors"

var (
	ErrBackendAlreadyExists = errors.New("backend already created")

	ErrReadConfigurationFailure = errors.New("failed to read configuration")
	ErrLoadConfigurationFailure = errors.New("failed to load configuration")

	ErrMissingToken = errors.New("missing token in session")
	ErrMissingUser  = errors.New("missing user in session")

	ErrMissingParameter = errors.New("missing parameter \"%s\" in request")
	ErrWelcomerMissing  = errors.New("bot is missing from server")
)
