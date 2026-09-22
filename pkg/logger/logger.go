package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the global zerolog logger with console writer or json output.
func Init(debug bool) {
	zerolog.TimeFieldFormat = time.RFC3339
	var output io.Writer = zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04:05",
	}

	level := zerolog.InfoLevel
	if debug {
		level = zerolog.DebugLevel
	}

	log.Logger = zerolog.New(output).Level(level).With().Timestamp().Logger()
}

// L returns the global logger instance.
func L() *zerolog.Logger {
	return &log.Logger
}
