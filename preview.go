package chunker

// PreviewRequest requests a preview from Chunker.
type PreviewRequest struct {
	Method string `json:"method"`
	Type string `json:"type"`
	Request
}

// NewPreviewRequest creates a new preview request.
func NewPreviewRequest() PreviewRequest {
	return PreviewRequest{
		Method:  "generate_preview",
		Request: NewRequest("auth"),
		Type: "flow",
	}
}