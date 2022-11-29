package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/rate"
	"github.com/harlow/go-micro-services/tracing"
	"github.com/harlow/go-micro-services/tune"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"time"
)

func main() {
	tune.Init()
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Caller().Logger()

	log.Info().Msg("Reading config...")
	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Error().Msgf("Got error while reading config: %v", err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	log.Info().Msgf("Read database URL: %v", result["RateMongoAddress"])
	log.Info().Msg("Initializing DB connection...")
	mongo_session := initializeDatabase(result["RateMongoAddress"])
	defer mongo_session.Close()
	log.Info().Msg("Successfull")

	log.Info().Msgf("Read profile memcashed address: %v", result["RateMemcAddress"])
	log.Info().Msg("Initializing Memcashed client...")
	memc_client := tune.NewMemCClient(result["RateMemcAddress"])
	log.Info().Msg("Successfull")

	serv_port, _ := strconv.Atoi(result["RatePort"])
	serv_ip := result["RateIP"]

	log.Info().Msgf("Read target port: %v", serv_port)
	log.Info().Msgf("Read consul address: %v", result["consulAddress"])
	log.Info().Msgf("Read jaeger address: %v", result["jaegerAddress"])
	var (
		// port       = flag.Int("port", 8084, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger server addr")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "rate", *jaegeraddr)
	tracer, err := tracing.Init("rate", *jaegeraddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing jaeger agent: %v", err)
	}
	log.Info().Msg("Jaeger agent initialized")

	log.Info().Msgf("Initializing consul agent [host: %v]...", *consuladdr)
	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	log.Info().Msg("Consul agent initialized")

	srv := &rate.Server{
		Tracer: tracer,
		// Port:     *port,
		Registry:     registry,
		Port:         serv_port,
		IpAddr:       serv_ip,
		MongoSession: mongo_session,
		MemcClient:   memc_client,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
