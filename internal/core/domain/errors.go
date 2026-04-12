package domain

import "errors"

// LLM / outbound generation errors (mapped to HTTP in the driving adapter).
var (
	ErrLLMQuotaOrRate = errors.New("llm: quota or rate limit")
	ErrLLMNoCapacity  = errors.New("llm: all keys exhausted or in cooldown")
)
