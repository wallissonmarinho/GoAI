package ports

import "context"

// TextCompletion is the outbound port for a single-turn text generation (e.g. Gemini).
type TextCompletion interface {
	GenerateText(ctx context.Context, userPrompt string) (string, error)
}
