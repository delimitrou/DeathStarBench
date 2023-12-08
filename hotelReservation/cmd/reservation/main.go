package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/delimitrou/DeathStarBench/hotelreservation/registry"
	"github.com/delimitrou/DeathStarBench/hotelreservation/services/reservation"
	"github.com/delimitrou/DeathStarBench/hotelreservation/tracing"
	"github.com/delimitrou/DeathStarBench/hotelreservation/tune"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	log.Info().Msg("Initializing DB connection...")
	mongoClient, mongoClose := initializeDatabase(result["ReserveMongoAddress"])
	defer mongoClose()

	log.Info().Msgf("Read profile memcashed address: %v", result["ReserveMemcAddress"])
	log.Info().Msg("Initializing Memcashed client...")
	memcClient := tune.NewMemCClient2(result["ReserveMemcAddress"])
	log.Info().Msg("Success")

	servPort, _ := strconv.Atoi(result["ReservePort"])
	servIP := result["ReserveIP"]

	var (
		jaegerAddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
		consulAddr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "reservation", *jaegerAddr)
	tracer, err := tracing.Init("reservation", *jaegerAddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing jaeger agent: %v", err)
	}
	log.Info().Msg("Jaeger agent initialized")

	log.Info().Msgf("Initializing consul agent [host: %v]...", *consulAddr)
	registry, err := registry.NewClient(*consulAddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing consul agent: %v", err)
	}
	log.Info().Msg("Consul agent initialized")

	srv := &reservation.Server{
		Tracer:      tracer,
		Registry:    registry,
		Port:        servPort,
		IpAddr:      servIP,
		MongoClient: mongoClient,
		MemcClient:  memcClient,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
