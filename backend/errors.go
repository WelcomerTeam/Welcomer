package backend

import "errors"

var (
	ErrBackendAlreadyExists = errors.New("backend already created")

	ErrReadConfigurationFailure = errors.New("failed to read configuration")
	ErrLoadConfigurationFailure = errors.New("failed to load configuration")

	ErrMissingToken = errors.New("missing token in session")
)
