package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/rate"
	"github.com/harlow/go-micro-services/tracing"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	"time"
)

func main() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	mongo_session := initializeDatabase(result["RateMongoAddress"])

	fmt.Printf("rate memc addr port = %s\n", result["RateMemcAddress"])
	memc_client := memcache.New(result["RateMemcAddress"])
	memc_client.Timeout = time.Second * 2
	memc_client.MaxIdleConns = 512

	defer mongo_session.Close()
	serv_port, _ := strconv.Atoi(result["RatePort"])
	serv_ip   := result["RateIP"]

	fmt.Printf("rate ip = %s, port = %d\n", serv_ip, serv_port)

	var (
		// port       = flag.Int("port", 8084, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["consulAddress"], "Jaeger server addr")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("rate", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &rate.Server{
		Tracer:   tracer,
		// Port:     *port,
		Registry: registry,
		Port:     serv_port,
		IpAddr:	  serv_ip,
		MongoSession: mongo_session,
		MemcClient: memc_client,
	}
	log.Fatal(srv.Run())
}
