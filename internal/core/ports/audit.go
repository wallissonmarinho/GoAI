package ports

import (
	"context"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
)

// AuditService is the application port used by HTTP (driving adapter).
type AuditService interface {
	AuditSeries(ctx context.Context, in domain.SeriesAuditRequest) (domain.SeriesAuditResponse, error)
	AuditRelease(ctx context.Context, in domain.ReleaseAuditRequest) (domain.ReleaseAuditResponse, error)
	PromptVersion() int
}
