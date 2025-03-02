package tests

import (
	"testing"

	"github.com/rs/zerolog"
)

type TLogWriter struct {
	t *testing.T
}

// Write implements the io.Writer interface for TLogWriter.
func (w *TLogWriter) Write(p []byte) (n int, err error) {
	w.t.Logf("%s", p)
	return len(p), nil
}

// Return a fake logger to satisfy functions that expect one
func NilLogger() *zerolog.Logger {
	logger := zerolog.New(nil)
	return &logger
}

// Return a logger that makes use of the T.Log method to enable debugging tests
func DebugLogger(t *testing.T) *zerolog.Logger {
	logger := zerolog.New(GetTLogWriter(t))
	return &logger
}

func GetTLogWriter(t *testing.T) *TLogWriter {
	return &TLogWriter{t: t}
}
