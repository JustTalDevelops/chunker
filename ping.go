package chunker

// PingRequest is a ping to Chunker to show them that the connection is still alive.
type PingRequest struct {
	Method string `json:"method"`
	Request
}

// NewPingRequest creates a new ping request.
func NewPingRequest() PingRequest {
	return PingRequest{
		Method:  "ping",
		Request: NewRequest("auth"),
	}
}