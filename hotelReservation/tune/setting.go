package tune

import (
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	defaultGCPercent        int    = 100
	defaultMemCTimeout      int    = 2
	defaultMemCMaxIdleConns int    = 512
	defaultLogLevel         string = "info"
)

func setGCPercent() {
	ratio := defaultGCPercent
	if val, ok := os.LookupEnv("GC"); ok {
		ratio, _ = strconv.Atoi(val)
	}

	debug.SetGCPercent(ratio)
	log.Info().Msgf("Tune: setGCPercent to %d", ratio)
}

func setLogLevel() {
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

func GetMemCTimeout() int {
	timeout := defaultMemCTimeout
	if val, ok := os.LookupEnv("MEMC_TIMEOUT"); ok {
		timeout, _ = strconv.Atoi(val)
	}
	log.Info().Msgf("Tune: GetMemCTimeout %d", timeout)
	return timeout
}

// Hack of memcache.New to avoid 'no server error' during running
func NewMemCClient(server ...string) *memcache.Client {
	ss := new(memcache.ServerList)
	err := ss.SetServers(server...)
	if err != nil {
		// Hack: panic early to avoid pod restart during running
		panic(err)
		//return nil, err
	} else {
		memc_client := memcache.NewFromSelector(ss)
		memc_client.Timeout = time.Second * time.Duration(GetMemCTimeout())
		memc_client.MaxIdleConns = defaultMemCMaxIdleConns
		return memc_client
	}
}

func NewMemCClient2(servers string) *memcache.Client {
	ss := new(memcache.ServerList)
	server_list := strings.Split(servers, ",")
	err := ss.SetServers(server_list...)
	if err != nil {
		// Hack: panic early to avoid pod restart during running
		panic(err)
		//return nil, err
	} else {
		memc_client := memcache.NewFromSelector(ss)
		memc_client.Timeout = time.Second * time.Duration(GetMemCTimeout())
		memc_client.MaxIdleConns = defaultMemCMaxIdleConns
		return memc_client
	}
}

func Init() {
	setLogLevel()
	setGCPercent()
}
