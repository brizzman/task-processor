package circuitbreaker

import (
	"errors"
	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"
	"testing"
	"time"

	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestBaseDecorator_ExecuteWithCB_Success(t *testing.T) {
	// Setup configuration and logger
	cfg := &config.Config{
		CircuitBreaker: config.CircuitBreaker{
			MaxRequests:         1,
			Timeout:             time.Second,
			Interval:            time.Second,
			ConsecutiveFailures: 2,
		},
	}
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}

	base := NewBaseDecorator(cfg, log, "test-component")
	base.AddCircuitBreaker("op", base.CreateSettings(cfg, "op"))

	// Execute a successful function through circuit breaker
	result, err := base.ExecuteWithCB("op", func() (any, error) {
		return 42, nil
	})

	// Assert the result is returned and no error occurred
	assert.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestBaseDecorator_ExecuteWithCB_Error_OpensCircuit(t *testing.T) {
	// Setup configuration with 0 consecutive failure threshold
	cfg := &config.Config{
		CircuitBreaker: config.CircuitBreaker{
			MaxRequests:         1,
			Timeout:             100 * time.Millisecond,
			Interval:            time.Millisecond,
			ConsecutiveFailures: 0,
		},
	}
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	base := NewBaseDecorator(cfg, log, "test-component")
	base.AddCircuitBreaker("op", base.CreateSettings(cfg, "op"))

	// First call fails → should trigger circuit breaker to OPEN
	_, err := base.ExecuteWithCB("op", func() (any, error) {
		return nil, errors.New("fail")
	})
	assert.Error(t, err)

	// Second call → CB is OPEN, should reject immediately with ErrOpenState
	_, err = base.ExecuteWithCB("op", func() (any, error) {
		return 42, nil
	})
	assert.ErrorIs(t, err, gobreaker.ErrOpenState)
}

func TestBaseDecorator_ExecuteWithCB_HalfOpenToClosed(t *testing.T) {
	cfg := &config.Config{
		CircuitBreaker: config.CircuitBreaker{
			MaxRequests:         1,
			Timeout:             100 * time.Millisecond, // short timeout to trigger Half-Open
			Interval:            time.Millisecond,
			ConsecutiveFailures: 0, // open immediately on first failure
		},
	}
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	base := NewBaseDecorator(cfg, log, "test-component")
	base.AddCircuitBreaker("op", base.CreateSettings(cfg, "op"))

	// 1. First call fails → circuit goes Open
	_, err := base.ExecuteWithCB("op", func() (any, error) {
		return nil, errors.New("fail")
	})
	assert.Error(t, err)

	// 2. Wait for timeout → circuit transitions to Half-Open
	time.Sleep(150 * time.Millisecond)

	// 3. Next call succeeds → circuit should move back to Closed
	_, err = base.ExecuteWithCB("op", func() (any, error) {
		return 42, nil
	})
	assert.NoError(t, err)

	// 4. Another call should also succeed (circuit now Closed again)
	_, err = base.ExecuteWithCB("op", func() (any, error) {
		return 99, nil
	})
	assert.NoError(t, err)
}

func TestBaseDecorator_ExecuteWithCB_HalfOpenToOpen(t *testing.T) {
	cfg := &config.Config{
		CircuitBreaker: config.CircuitBreaker{
			MaxRequests:         1,
			Timeout:             100 * time.Millisecond,
			Interval:            time.Millisecond,
			ConsecutiveFailures: 0,
		},
	}
	log := &logger.ZapLogger{Logger: zaptest.NewLogger(t)}
	base := NewBaseDecorator(cfg, log, "test-component")
	base.AddCircuitBreaker("op", base.CreateSettings(cfg, "op"))

	// 1. First call fails → breaker goes to Open
	_, err := base.ExecuteWithCB("op", func() (any, error) {
		return nil, errors.New("fail")
	})
	assert.Error(t, err)

	// 2. Wait for timeout → breaker moves to Half-Open
	time.Sleep(150 * time.Millisecond)

	// 3. Trial call fails again → breaker transitions Half-Open → Open
	_, err = base.ExecuteWithCB("op", func() (any, error) {
		return nil, errors.New("fail again")
	})
	assert.Error(t, err)

	// 4. Another call should be rejected immediately (breaker still Open)
	_, err = base.ExecuteWithCB("op", func() (any, error) {
		return 123, nil
	})
	assert.ErrorIs(t, err, gobreaker.ErrOpenState)
}