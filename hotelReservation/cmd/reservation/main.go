package main

import (
	"context"
	"encoding/json"
	"flag"
	"io"
	"os"

	"strconv"

	"github.com/harlow/go-micro-services/services/reservation"
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

	byteValue, _ := io.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	log.Info().Msgf("Read database URL: %v", result["ReserveMongoAddress"])
	log.Info().Msg("Initializing DB connection...")
	ctx := context.Background()
	mongo_client := initializeDatabase(ctx, result["ReserveMongoAddress"])
	defer mongo_client.Disconnect(ctx)
	log.Info().Msg("Successfull")

	log.Info().Msgf("Read profile memcashed address: %v", result["ReserveMemcAddress"])
	log.Info().Msg("Initializing Memcashed client...")
	memc_client := tune.NewPiCoPMemCClient(result["ReserveMemcAddress"])
	defer memc_client.Close()
	log.Info().Msg("Successfull")

	serv_port, _ := strconv.Atoi(result["ReservePort"])
	serv_ip := result["ReserveIP"]
	log.Info().Msgf("Read target port: %v", serv_port)
	log.Info().Msgf("Read consul address: %v", result["consulAddress"])
	log.Info().Msgf("Read jaeger address: %v", result["jaegerAddress"])

	var (
		// port       = flag.Int("port", 8087, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger server addr")
	)
	flag.Parse()

	log.Info().Msgf("Initializing jaeger agent [service name: %v | host: %v]...", "reservation", *jaegeraddr)
	tracer, err := tracing.Init("reservation", *jaegeraddr)
	if err != nil {
		log.Panic().Msgf("Got error while initializing jaeger agent: %v", err)
	}
	log.Info().Msg("Jaeger agent initialized")

	srv := &reservation.Server{
		Tracer: tracer,
		// Port:     *port,
		Port:        serv_port,
		IpAddr:      serv_ip,
		MongoClient: mongo_client,
		MemcClient:  memc_client,
	}

	log.Info().Msg("Starting server...")
	log.Fatal().Msg(srv.Run().Error())
}
