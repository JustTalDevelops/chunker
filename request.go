package chunker

import (
	"github.com/google/uuid"
)

// Request is a Chunker request. It simply contains a request ID, which is a random UUID.
type Request struct {
	RequestId string `json:"requestId"`
	Type string `json:"type"`
}

// NewRequest returns a new Chunker request.
func NewRequest(requestType string) Request {
	return Request{RequestId: uuid.NewString(), Type: requestType}
}