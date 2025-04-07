package welcomer

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func SetupLogger(loggingLevel string) io.Writer {
	// Setup Logger

	level, err := zerolog.ParseLevel(loggingLevel)
	if err != nil {
		panic(fmt.Errorf(`failed to parse loggingLevel. zerolog.ParseLevel(%s): %w`, loggingLevel, err))
	}

	zerolog.SetGlobalLevel(level)

	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.Stamp,
	}

	Logger = zerolog.New(writer).With().Timestamp().Logger()
	Logger.Info().Msg("Logging configured")

	return writer
}
