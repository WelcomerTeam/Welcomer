package backend

import (
	"errors"
)

var (
	ErrBackendAlreadyExists = errors.New("backend already created")

	ErrMissingToken = errors.New("missing token in session")
	ErrMissingUser  = errors.New("missing user in session")

	ErrMissingParameter = errors.New("missing parameter \"%s\" in request")
	ErrInvalidParameter = errors.New("invalid parameter \"%s\" in request")
	ErrWelcomerMissing  = errors.New("bot is missing from server")
	ErrEnsureFailure    = errors.New("failed to ensure guild")

	ErrOAuthFailure = errors.New("issue checking oauth2 token")

	ErrMembershipAlreadyInUse = errors.New("membership is already in use")
	ErrMembershipInvalid      = errors.New("membership is invalid")
	ErrMembershipExpired      = errors.New("membership has expired")
	ErrMembershipNotInUse     = errors.New("membership is not in use")
	ErrUnhandledMembership    = errors.New("membership type is not handled")
)

// HTTP errors.
var (
	ErrInvalidContentType = errors.New("content type not accepted")
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

	ErrBackgroundTooLarge = errors.New("background size is too large")
	ErrFileSizeTooLarge   = errors.New("this file has an image resolution that is too high")
	ErrFileNotSupported   = errors.New("this file format is not supported")
	ErrConversionFailed   = errors.New("failed to convert background")

	ErrCannotUseCustomBackgrounds = errors.New("you cannot upload custom utils.backgrounds")

	ErrStringTooLong = errors.New("string is too long")
	ErrListTooLong   = errors.New("list is too long")
)

// Borderwall errors.
var (
	ErrBorderwallRequestAlreadyVerified = errors.New("borderwall request already verified")
	ErrBorderwallInvalidKey             = errors.New("invalid key")
	ErrBorderwallUserInvalid            = errors.New("user is not the owner of this request")
	ErrRecaptchaValidationFailed        = errors.New("reCAPTCHA validation failed")
	ErrInsecureUser                     = errors.New("failed to verify your request. Please disable any proxy or VPN and try again")
)
