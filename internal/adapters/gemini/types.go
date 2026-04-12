package gemini

// generateContentRequest matches Gemini REST v1beta (simplified).
type generateContentRequest struct {
	Contents         []contentBlock `json:"contents"`
	GenerationConfig genConfig      `json:"generationConfig"`
}

type contentBlock struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type genConfig struct {
	Temperature      float32 `json:"temperature,omitempty"`
	ResponseMIMEType string  `json:"responseMimeType,omitempty"`
}

type generateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []part `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}
