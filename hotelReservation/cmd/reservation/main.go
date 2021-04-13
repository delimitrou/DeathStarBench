package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/reservation"
	"github.com/harlow/go-micro-services/tracing"
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
	serv_port, _ := strconv.Atoi(result["ReservePort"])
	serv_ip := ""
	reserve_mongo_addr := ""
	reserve_memc_addr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s"{
		reserve_mongo_addr = "mongodb-reserve:"+strings.Split(result["ReserveMongoAddress"], ":")[1]
		reserve_memc_addr = "memcached-reserve:"+strings.Split(result["ReserveMemcAddress"], ":")[1]
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
		reserve_mongo_addr = result["ReserveMongoAddress"]
		reserve_memc_addr = result["ReserveMemcAddress"]
		serv_ip = result["ReserveIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongo_session := initializeDatabase(reserve_mongo_addr)
	defer mongo_session.Close()

	fmt.Printf("reservation memc addr port = %s\n", result["ReserveMemcAddress"])
	memc_client := memcache.New(reserve_memc_addr)
	memc_client.Timeout = time.Second * 2
	memc_client.MaxIdleConns = 512



	fmt.Printf("reservation ip = %s, port = %d\n", serv_ip, serv_port)



	tracer, err := tracing.Init("reservation", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &reservation.Server{
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