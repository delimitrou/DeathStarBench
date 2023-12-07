package tune

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	defaultLogLevel string = "info"
)

func setLogLevel()  {
	logLevel := defaultLogLevel
	if val, ok := os.LookupEnv("LOG_LEVEL"); ok {
		logLevel = val
	}
        switch logLevel {
        case "", "ERROR", "error": // If env is unset, set level to ERROR.
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
        case "WARNING", "warning":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
        case "DEBUG", "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
        case "INFO", "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
        case "TRACE", "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	default: // Set default log level to info
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
        }

	log.Info().Msgf("Set global log level: %s", logLevel)
}

func Init() {
	setLogLevel()
}
