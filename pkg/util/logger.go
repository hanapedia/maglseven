package util

// Logger interface with methods matching slog's signature
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
}
