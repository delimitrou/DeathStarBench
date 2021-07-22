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
	"github.com/harlow/go-micro-services/services/profile"
	"github.com/harlow/go-micro-services/tracing"

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

	mongo_session := initializeDatabase(result["ProfileMongoAddress"])
	defer mongo_session.Close()

	fmt.Printf("profile memc addr port = %s\n", result["ProfileMemcAddress"])
	memc_client := memcache.New(result["ProfileMemcAddress"])
	memc_client.Timeout = time.Second * 2
	memc_client.MaxIdleConns = 512

	serv_port, _ := strconv.Atoi(result["ProfilePort"])
	serv_ip   := result["ProfileIP"]

	fmt.Printf("profile ip = %s, port = %d\n", serv_ip, serv_port)

	var (
		// port       = flag.Int("port", 8081, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["consulAddress"], "Jaeger server addr")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	flag.Parse()

	tracer, err := tracing.Init("profile", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := profile.Server{
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
