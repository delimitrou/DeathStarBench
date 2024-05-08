package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/review"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tracing"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tune"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"time"
	// "github.com/bradfitz/gomemcache/memcache"
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

	log.Info().Msgf("Read database URL: %v", result["ReviewMongoAddress"])
	log.Info().Msg("Initializing DB connection...")
	mongo_session, mongoClose := initializeDatabase(result["ReviewMongoAddress"])
	defer mongoClose()
	log.Info().Msg("Successfull")

	log.Info().Msgf("Read review memcashed address: %v", result["ReviewMemcAddress"])
	log.Info().Msg("Initializing Memcashed client...")
	// memc_client := memcache.New(result["ReviewMemcAddress"])
	// memc_client.Timeout = time.Second * 2
	// memc_client.MaxIdleConns = 512
	memc_client := tune.NewMemCClient2(result["ReviewMemcAddress"])
	log.Info().Msg("Successfull")

	serv_port, _ := strconv.Atoi(result["ReviewPort"])
	serv_ip := result["ReviewIP"]
	log.Info().Msgf("Read target port: %v", serv_port)
	log.Info().Msgf("Read consul address: %v", result["consulAddress"])
	log.Info().Msgf("Read jaeger address: %v", result["jaegerAddress"])

	var (
		// port       = flag.Int("port", 8081, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger server addr")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "review", *jaegeraddr)
	tracer, err := tracing.Init("review", *jaegeraddr)
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

	srv := review.Server{
		Tracer: tracer,
		// Port:     *port,
		Registry:    registry,
		Port:        serv_port,
		IpAddr:      serv_ip,
		MongoClient: mongo_session,
		MemcClient:  memc_client,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
