// Package contextkey provides context key types for request-scoped values.
package contextkey

import (
	"context"

	"github.com/google/uuid"
)

// Key types for type-safe context values.
type userIDKey struct{}
type correlationIDKey struct{}

// WithUserID adds user ID to context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey{}, userID)
}

// UserID extracts user ID from context.
func UserID(ctx context.Context) string {
	if id, ok := ctx.Value(userIDKey{}).(string); ok {
		return id
	}
	return ""
}

// WithCorrelationID adds correlation ID to context (or generates one if empty).
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		correlationID = uuid.New().String()
	}
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

// CorrelationID extracts correlation ID from context.
func CorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
		return id
	}
	return ""
}
