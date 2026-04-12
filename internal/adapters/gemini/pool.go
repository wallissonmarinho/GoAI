package gemini

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
)

// KeyPool rotates Gemini API keys and applies short cooldowns after quota errors.
type KeyPool struct {
	mu          sync.Mutex
	keys        []string
	cooldown    map[int]time.Time // index -> UTC until when key must not be used
	cooldownDur time.Duration
}

// NewKeyPool parses comma-separated API keys (non-empty after trim).
func NewKeyPool(csv string, cooldown time.Duration) *KeyPool {
	if cooldown <= 0 {
		cooldown = 60 * time.Second
	}
	var keys []string
	for _, s := range strings.Split(csv, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			keys = append(keys, s)
		}
	}
	return &KeyPool{
		keys:        keys,
		cooldown:    make(map[int]time.Time),
		cooldownDur: cooldown,
	}
}

// Len returns the number of configured keys.
func (p *KeyPool) Len() int {
	if p == nil {
		return 0
	}
	return len(p.keys)
}

// SnapshotKeys returns a copy of configured API keys for server-side diagnostics only.
func (p *KeyPool) SnapshotKeys() []string {
	if p == nil || len(p.keys) == 0 {
		return nil
	}
	out := make([]string, len(p.keys))
	copy(out, p.keys)
	return out
}

// WithKey tries each key in order; on domain.ErrLLMQuotaOrRate the key gets a cooldown and the next is tried.
func (p *KeyPool) WithKey(now time.Time, fn func(apiKey string) error) error {
	if p == nil || len(p.keys) == 0 {
		return domain.ErrLLMNoCapacity
	}
	var lastErr error
	for i, k := range p.keys {
		p.mu.Lock()
		if until, ok := p.cooldown[i]; ok && now.Before(until) {
			p.mu.Unlock()
			continue
		}
		p.mu.Unlock()

		err := fn(k)
		if err == nil {
			return nil
		}
		lastErr = err
		if errors.Is(err, domain.ErrLLMQuotaOrRate) {
			p.mu.Lock()
			p.cooldown[i] = now.Add(p.cooldownDur)
			p.mu.Unlock()
			continue
		}
		return err
	}
	if lastErr != nil {
		return lastErr
	}
	return domain.ErrLLMNoCapacity
}
