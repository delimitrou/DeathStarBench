package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

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

	serv_port, _ := strconv.Atoi(result["ProfilePort"])
	serv_ip := ""
	profile_mongo_addr := ""
	profile_memc_addr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s"{
		profile_mongo_addr = "mongodb-profile:"+strings.Split(result["ProfileMongoAddress"], ":")[1]
		profile_memc_addr = "memcached-profile:"+strings.Split(result["ProfileMemcAddress"], ":")[1]
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
		profile_mongo_addr = result["ProfileMongoAddress"]
		profile_memc_addr = result["ProfileMemcAddress"]
		serv_ip = result["ProfileIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongo_session := initializeDatabase(profile_mongo_addr)
	defer mongo_session.Close()

	fmt.Printf("profile memc addr port = %s\n", profile_memc_addr)
	memc_client := memcache.New(profile_memc_addr)
	memc_client.Timeout = time.Second * 2
	memc_client.MaxIdleConns = 512

	fmt.Printf("profile ip = %s, port = %d\n", serv_ip, serv_port)

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
