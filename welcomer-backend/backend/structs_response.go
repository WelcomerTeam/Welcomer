package backend

// BaseResponse represents the base response sent to a client.
type BaseResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error string      `json:"error,omitempty"`
	Ok    bool        `json:"ok"`
}
