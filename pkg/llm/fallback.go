package llm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// ModelRoute defines an endpoint route in the fallback chain.
type ModelRoute struct {
	Name       string
	Client     LLMClient
	Model      string
	MaxRetries int
}

// FallbackRouter manages a multi-model fallback chain with automatic retries.
type FallbackRouter struct {
	routes []ModelRoute
}

// NewFallbackRouter creates a new FallbackRouter.
func NewFallbackRouter(routes []ModelRoute) *FallbackRouter {
	return &FallbackRouter{routes: routes}
}

// ProviderName returns the provider identifier.
func (r *FallbackRouter) ProviderName() string {
	return "fallback_router"
}

// isRetryableError checks if an error can be retried.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "status 429") ||
		strings.Contains(msg, "status 500") ||
		strings.Contains(msg, "status 502") ||
		strings.Contains(msg, "status 503") ||
		strings.Contains(msg, "status 504") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "temporary")
}

// Generate attempts generation through the fallback chain with retries.
func (r *FallbackRouter) Generate(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if len(r.routes) == 0 {
		return nil, fmt.Errorf("no model routes configured in fallback router")
	}

	var allErrors []string

	for routeIdx, route := range r.routes {
		retries := route.MaxRetries
		if retries <= 0 {
			retries = 2
		}

		routeReq := req
		if route.Model != "" {
			routeReq.Model = route.Model
		}

		backoff := 500 * time.Millisecond
		var lastErr error

		for attempt := 1; attempt <= retries; attempt++ {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			resp, err := route.Client.Generate(ctx, routeReq)
			if err == nil {
				if routeIdx > 0 {
					log.Info().
						Str("route", route.Name).
						Str("model", route.Model).
						Msg("request successfully handled by backup model")
				}
				return resp, nil
			}

			lastErr = err
			if !isRetryableError(err) {
				break
			}

			time.Sleep(backoff)
			backoff *= 2
			if backoff > 3*time.Second {
				backoff = 3 * time.Second
			}
		}

		errSummary := fmt.Sprintf("[%s (%s)]: %v", route.Name, route.Model, lastErr)
		allErrors = append(allErrors, errSummary)
		log.Warn().
			Str("route", route.Name).
			Int("retries", retries).
			Err(lastErr).
			Msg("route failed after retries, failing over to next model")
	}

	return nil, fmt.Errorf("all routes in fallback chain failed: %s", strings.Join(allErrors, "; "))
}

// Stream attempts streaming through the fallback chain.
func (r *FallbackRouter) Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error) {
	if len(r.routes) == 0 {
		return nil, fmt.Errorf("no model routes configured in fallback router")
	}

	for _, route := range r.routes {
		routeReq := req
		if route.Model != "" {
			routeReq.Model = route.Model
		}

		ch, err := route.Client.Stream(ctx, routeReq)
		if err == nil {
			return ch, nil
		}

		log.Warn().
			Str("route", route.Name).
			Err(err).
			Msg("stream initialization failed on route, trying next route")
	}

	return nil, fmt.Errorf("failed to initialize stream on any route in fallback chain")
}
