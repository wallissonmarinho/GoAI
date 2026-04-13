package ports

import "context"

// URLExistenceChecker validates whether a public URL resolves.
type URLExistenceChecker interface {
	Exists(ctx context.Context, rawURL string) (bool, error)
}
