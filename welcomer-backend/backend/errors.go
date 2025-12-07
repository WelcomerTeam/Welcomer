package backend

import (
	"fmt"
	"runtime"
	"strconv"
)

type BackendError struct {
	Message string
	Code    int
}

func (e BackendError) Error() string {
	return strconv.Itoa(e.Code) + ": " + e.Message
}

func NewErrorWithCode(code int, message string) error {
	return BackendError{
		Message: message,
		Code:    code,
	}
}

// HTTP errors.
var (
	ErrInvalidContentType = NewErrorWithCode(406, "content type not accepted")
)

var (
	ErrBackendAlreadyExists = NewErrorWithCode(10000, "backend already created")
	ErrMissingToken         = NewErrorWithCode(10001, "missing token in session")
	ErrMissingUser          = NewErrorWithCode(10002, "missing user in session")
	ErrMissingParameter     = NewErrorWithCode(10003, "missing parameter in request")
	ErrInvalidParameter     = NewErrorWithCode(10004, "invalid parameter in request")
	ErrWelcomerMissing      = NewErrorWithCode(10005, "bot is missing from server")
	ErrEnsureFailure        = NewErrorWithCode(10006, "failed to ensure guild")
	ErrOAuthFailure         = NewErrorWithCode(10007, "issue checking oauth2 token")
	ErrInvalidPermissions   = NewErrorWithCode(10008, "invalid permissions for this endpoint")
)

func NewMissingParameterError(parameter string) error {
	return NewErrorWithCode(10003, fmt.Sprintf("missing parameter \"%s\" in request", parameter))
}

func NewInvalidParameterError(parameter string) error {
	return NewErrorWithCode(10004, fmt.Sprintf("invalid parameter \"%s\" in request", parameter))
}

func NewGenericErrorWithLineNumber() error {
	// Get the line number of the caller.
	_, _, lineNo, ok := runtime.Caller(1)

	if ok {
		return NewErrorWithCode(11000, fmt.Sprintf("general error: l%d", lineNo))
	}

	return ErrGeneralError
}

// Validation errors.
var (
	ErrGeneralError = NewErrorWithCode(11000, "general error")

	ErrRequired                   = NewErrorWithCode(11001, "this field is required")
	ErrChannelInvalid             = NewErrorWithCode(11002, "this channel does not exist")
	ErrInvalidJSON                = NewErrorWithCode(11003, "invalid json")
	ErrInvalidColour              = NewErrorWithCode(11004, "colour format is not recognised")
	ErrInvalidBackground          = NewErrorWithCode(11005, "invalid background")
	ErrInvalidImageAlignment      = NewErrorWithCode(11006, "image alignment is not recognised")
	ErrInvalidImageTheme          = NewErrorWithCode(11007, "image theme is not recognised")
	ErrInvalidProfileBorderType   = NewErrorWithCode(11008, "profile border type is not recognised")
	ErrBackgroundTooLarge         = NewErrorWithCode(11009, "background size is too large")
	ErrFileSizeTooLarge           = NewErrorWithCode(11010, "this file has an image resolution that is too high")
	ErrFileNotSupported           = NewErrorWithCode(11011, "this file format is not supported")
	ErrConversionFailed           = NewErrorWithCode(11012, "failed to convert background")
	ErrCannotUseCustomBackgrounds = NewErrorWithCode(11013, "you cannot upload custom welcomer.backgrounds")
	ErrStringTooLong              = NewErrorWithCode(11014, "string is too long")
	ErrListTooLong                = NewErrorWithCode(11015, "list is too long")

	ErrTooManyLayers     = NewErrorWithCode(11016, "too many layers in custom image")
	ErrDimensionTooLarge = NewErrorWithCode(11017, "custom image dimensions are too large")
	ErrDimensionTooSmall = NewErrorWithCode(11018, "custom image dimensions are too small")
)

// Borderwall errors.
var (
	ErrBorderwallRequestAlreadyVerified = NewErrorWithCode(12000, "borderwall request already verified")
	ErrBorderwallInvalidKey             = NewErrorWithCode(12001, "invalid key")
	ErrBorderwallUserInvalid            = NewErrorWithCode(12002, "user is not the owner of this request")

	ErrRecaptchaValidationFailed = NewErrorWithCode(12003, "reCAPTCHA validation failed")
	ErrInsecureUser              = NewErrorWithCode(12004, "failed to verify your request. Please disable any proxy or VPN and try again")
)

// Billing errors.
var (
	ErrInvalidSKU                 = NewErrorWithCode(13000, "invalid sku")
	ErrCurrencyNotAvailable       = NewErrorWithCode(13001, "currency is not available")
	ErrCurrencyNotAvailableForSKU = NewErrorWithCode(13002, "no currency pricing is available for this sku")
	ErrInvalidCost                = NewErrorWithCode(13003, "invalid cost")
	ErrCreateOrderFailed          = NewErrorWithCode(13004, "failed to create paypal order")
	ErrGetOrderValidationFailed   = NewErrorWithCode(13005, "failed to validate paypal order")
	ErrCaptureOrderFailed         = NewErrorWithCode(13006, "failed to capture paypal order")
	ErrCreateTransactionFailed    = NewErrorWithCode(13007, "failed to create transaction")
	ErrCreateMembershipFailed     = NewErrorWithCode(13008, "failed to create membership")
	ErrWebhookValidationFailed    = NewErrorWithCode(13009, "failed to validate webhook message")
	ErrCannotTransferMembership   = NewErrorWithCode(13010, "cannot transfer membership")
)

// Custom bot errors.
var (
	ErrInvalidToken          = NewErrorWithCode(13100, "invalid token")
	ErrCustomBotLimitReached = NewErrorWithCode(13101, "custom bot limit reached")
)
