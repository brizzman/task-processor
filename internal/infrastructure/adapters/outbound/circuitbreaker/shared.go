package circuitbreaker

import (
	"fmt"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"

	"task-processor/internal/infrastructure/config"
	"task-processor/internal/infrastructure/shared/logger"
)

type BaseDecorator struct {
	circuitBreakers map[string]*gobreaker.CircuitBreaker
	logger          *logger.Logger
	name            string
}

func NewBaseDecorator(cfg *config.Config, logger *logger.Logger, name string) *BaseDecorator {
	return &BaseDecorator{
		circuitBreakers: make(map[string]*gobreaker.CircuitBreaker),
		logger:          logger,
		name:            name,
	}
}

func (d *BaseDecorator) AddCircuitBreaker(operation string, settings gobreaker.Settings) {
	d.circuitBreakers[operation] = gobreaker.NewCircuitBreaker(settings)
}

func (d *BaseDecorator) ExecuteWithCB(operation string, fn func() (any, error)) (any, error) {
	cb, exists := d.circuitBreakers[operation]
	if !exists {
		return nil, fmt.Errorf("circuit breaker for operation %s not found", operation)
	}

	d.logger.Debug("circuit breaker executing operation",
		zap.String("name", d.name),
		zap.String("operation", operation),
		zap.String("state", cb.State().String()))

	start := time.Now()
	result, err := cb.Execute(fn)
	duration := time.Since(start)

	if err != nil {
		d.handleError(operation, err, duration, cb.State())
		return nil, err
	}

	d.logger.Debug("circuit breaker operation succeeded",
		zap.String("name", d.name),
		zap.String("operation", operation),
		zap.String("state", cb.State().String()),
		zap.Duration("duration", duration))

	return result, nil
}

func (d *BaseDecorator) handleError(operation string, err error, duration time.Duration, state gobreaker.State) {
	switch err {
	case gobreaker.ErrOpenState:
		d.logger.Warn("circuit breaker rejected request - open state",
			zap.String("name", d.name),
			zap.String("operation", operation),
			zap.String("state", "open"),
			zap.Duration("duration", duration))
	case gobreaker.ErrTooManyRequests:
		d.logger.Warn("circuit breaker rejected request - too many requests in half-open state",
			zap.String("name", d.name),
			zap.String("operation", operation),
			zap.String("state", "half-open"),
			zap.Duration("duration", duration))
	default:
		d.logger.Error("circuit breaker operation failed",
			zap.String("name", d.name),
			zap.String("operation", operation),
			zap.String("state", state.String()),
			zap.Error(err),
			zap.Duration("duration", duration))
	}
}

func (d *BaseDecorator) CreateSettings(cfg *config.Config, operation string) gobreaker.Settings {
	return gobreaker.Settings{
		Name:        fmt.Sprintf("%s-%s", d.name, operation),
		MaxRequests: cfg.CircuitBreaker.MaxRequests,
		Interval:    cfg.CircuitBreaker.Interval,
		Timeout:     cfg.CircuitBreaker.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures > cfg.CircuitBreaker.ConsecutiveFailures
		},
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			d.logger.Info("circuit breaker state transition",
				zap.String("component", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()))
		},
	}
}