package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/geo"
	"github.com/harlow/go-micro-services/tracing"
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

	mongo_session := initializeDatabase(result["GeoMongoAddress"])
	defer mongo_session.Close()
	serv_port, _ := strconv.Atoi(result["GeoPort"])
	serv_ip   := result["GeoIP"]

	fmt.Printf("geo ip = %s, port = %d\n", serv_ip, serv_port)
	
	var (
		// port       = flag.Int("port", 8083, "Server port")
		jaegeraddr = flag.String("jaegeraddr", result["consulAddress"], "Jaeger address")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("geo", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &geo.Server{
		// Port:     *port,
		Port:     serv_port,
		IpAddr:	  serv_ip,
		Tracer:   tracer,
		Registry: registry,
		MongoSession: mongo_session,
	}
	log.Fatal(srv.Run())
}
