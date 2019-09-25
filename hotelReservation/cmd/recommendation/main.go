package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/recommendation"
	"github.com/harlow/go-micro-services/tracing"
	"strconv"
	// "github.com/bradfitz/gomemcache/memcache"
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

	mongo_session := initializeDatabase(result["RecommendMongoAddress"])
	defer mongo_session.Close()

	serv_port, _ := strconv.Atoi(result["RecommendPort"])
	serv_ip   := result["RecommendIP"]

	fmt.Printf("recommendation ip = %s, port = %d\n", serv_ip, serv_port)

	var (
		// port       = flag.Int("port", 8085, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["consulAddress"], "Jaeger server addr")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("recommendation", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &recommendation.Server{
		Tracer:   tracer,
		// Port:     *port,
		Registry: registry,
		Port:     serv_port,
		IpAddr:	  serv_ip,
		MongoSession: mongo_session,
	}
	log.Fatal(srv.Run())
}
