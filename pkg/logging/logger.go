package logging

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// Takes a log level as string and converts it to a zerolog.Level interface.
// If the string is not a valid input it will return zerolog.InfoLevel
func GetLogLevel(level string) zerolog.Level {
	levels := map[string]zerolog.Level{
		"trace": zerolog.TraceLevel,
		"debug": zerolog.DebugLevel,
		"info":  zerolog.InfoLevel,
		"warn":  zerolog.WarnLevel,
		"error": zerolog.ErrorLevel,
		"fatal": zerolog.FatalLevel,
		"panic": zerolog.PanicLevel,
	}
	logLevel, valid := levels[level]
	if !valid {
		return zerolog.InfoLevel
	}
	return logLevel
}

// Returns a pointer to a new log file with the specified path.
// Remember to call file.Close() when finished writing to the log file
func GetLogFile(path string) (*os.File, error) {
	logPath := filepath.Join(path, "server.log")
	file, err := os.OpenFile(
		logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0663,
	)
	if err != nil {
		return nil, errors.Wrap(err, "os.OpenFile")
	}
	return file, nil
}

// Get a pointer to a new zerolog.Logger with the specified level and output
// Can provide a file, writer or both. Must provide at least one of the two
func GetLogger(
	logLevel zerolog.Level,
	w io.Writer,
	logFile *os.File,
	logDir string,
) (*zerolog.Logger, error) {
	if w == nil && logFile == nil {
		return nil, errors.New("No Writer provided for log output.")
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	var consoleWriter zerolog.ConsoleWriter
	if w != nil {
		consoleWriter = zerolog.ConsoleWriter{Out: w}
	}

	var output io.Writer
	if logFile != nil {
		if w != nil {
			output = zerolog.MultiLevelWriter(logFile, consoleWriter)
		} else {
			output = logFile
		}
	} else {
		output = consoleWriter
	}
	logger := zerolog.New(output).
		With().
		Timestamp().
		Logger().
		Level(logLevel)

	return &logger, nil
}
