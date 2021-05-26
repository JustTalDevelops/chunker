package chunker

// LoginRequest is a request to login to Chunker.
type LoginRequest struct {
	Method string `json:"method"`
	UUID string `json:"uuid"`
	Request
}

// NewLoginRequest creates a new login request from a world ID.
func NewLoginRequest(worldId string) LoginRequest {
	return LoginRequest{
		Method:  "login",
		UUID:    worldId,
		Request: NewRequest("auth"),
	}
}