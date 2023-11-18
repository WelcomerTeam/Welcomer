package backend

// BaseResponse represents the base response sent to a client.
type BaseResponse struct {
	Ok    bool        `json:"ok"`
	Error string      `json:"error,omitempty"`
	Data  interface{} `json:"data,omitempty"`
}
