package backend

import "errors"

// BaseResponse represents the base response sent to a client.
type BaseResponse struct {
	Ok    bool   `json:"ok"`
	Code  int    `json:"code,omitempty"`
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func NewBaseResponse(err error, data any) BaseResponse {
	var code int

	var errString string

	if err != nil {
		var errWithCode BackendError
		if errors.As(err, &errWithCode) {
			code = errWithCode.Code
		}

		errString = err.Error()
	}

	return BaseResponse{
		Ok:    errString == "",
		Data:  data,
		Code:  code,
		Error: errString,
	}
}
