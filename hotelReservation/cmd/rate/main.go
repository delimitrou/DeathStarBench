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
	"net"
	"os"
	"strconv"
	"strings"

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
	
	serv_port, _ := strconv.Atoi(result["RatePort"])
	serv_ip := ""
	rate_mongo_addr := ""
	rate_memc_addr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s"{
		rate_mongo_addr = "mongodb-rate:"+strings.Split(result["RateMongoAddress"], ":")[1]
		rate_memc_addr = "memcached-rate:"+strings.Split(result["RateMemcAddress"], ":")[1]
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					serv_ip = ipnet.IP.String()

				}
			}
		}
		*jaegeraddr = "jaeger:"+strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		rate_mongo_addr = result["RateMongoAddress"]
		rate_memc_addr = result["RateMemcAddress"]
		serv_ip = result["RateIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()
	

	mongo_session := initializeDatabase(rate_mongo_addr)

	fmt.Printf("rate memc addr port = %s\n", rate_memc_addr)
	memc_client := memcache.New(rate_memc_addr)
	memc_client.Timeout = time.Second * 2
	memc_client.MaxIdleConns = 512

	defer mongo_session.Close()

	fmt.Printf("rate ip = %s, port = %d\n", serv_ip, serv_port)



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
