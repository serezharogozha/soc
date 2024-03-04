package log

import (
	"github.com/rs/zerolog"
	"os"
)

type Hook struct{}

func InitLogger() *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &log
}
