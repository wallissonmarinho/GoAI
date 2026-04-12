package gemini

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
)

func TestKeyPool_WithKey_skipsCooldown(t *testing.T) {
	p := NewKeyPool("a,b", time.Hour)
	var calls int
	now := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	err := p.WithKey(now, func(apiKey string) error {
		calls++
		if apiKey == "a" {
			return domain.ErrLLMQuotaOrRate
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, calls, "same request should try a then b")

	calls = 0
	err = p.WithKey(now.Add(time.Second), func(apiKey string) error {
		calls++
		require.Equal(t, "b", apiKey, "a still on cooldown")
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, calls)
}

func TestKeyPool_WithKey_nonQuotaStops(t *testing.T) {
	p := NewKeyPool("only", time.Minute)
	err := p.WithKey(time.Now(), func(apiKey string) error {
		return errors.New("bad request")
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "bad request")
}
