// Package logger defines the logging interface for the application.
package logger

import "context"

// Logger defines the interface for structured logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

// Field represents a structured log field.
type Field struct {
	Key   string
	Value interface{}
}

func String(key, value string) Field           { return Field{Key: key, Value: value} }
func Strings(key string, value []string) Field { return Field{Key: key, Value: value} }
func Int(key string, value int) Field          { return Field{Key: key, Value: value} }
func Int64(key string, value int64) Field      { return Field{Key: key, Value: value} }
func Float64(key string, value float64) Field  { return Field{Key: key, Value: value} }
func Bool(key string, value bool) Field        { return Field{Key: key, Value: value} }
func Error(err error) Field                    { return Field{Key: "error", Value: err} }
func Any(key string, value interface{}) Field  { return Field{Key: key, Value: value} }

// Common field keys.
const (
	FieldCorrelationID = "correlation_id"
	FieldUserID        = "user_id"
	FieldMethod        = "method"
	FieldPath          = "path"
	FieldStatus        = "status"
	FieldDuration      = "duration_ms"
	FieldProvider      = "provider"
	FieldCurrency      = "currency"
	FieldComponent     = "component"
	FieldOperation     = "operation"
	FieldFromCurrency  = "from_currency"
	FieldToCurrency    = "to_currency"
	FieldRate          = "rate"
)
