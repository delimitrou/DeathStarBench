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

	serv_port, _ := strconv.Atoi(result["RecommendPort"])
	serv_ip := ""
	recommendation_mongo_addr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s"{
		recommendation_mongo_addr = "mongodb-recommendation:"+strings.Split(result["RecommendMongoAddress"], ":")[1]
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					serv_ip = ipnet.IP.String()

				}
			}
		}
		*jaegeraddr = "jaeger:" + strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]
	} else {
		recommendation_mongo_addr = result["RecommendMongoAddress"]
		serv_ip = result["RecommendIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongo_session := initializeDatabase(recommendation_mongo_addr)
	defer mongo_session.Close()



	fmt.Printf("recommendation ip = %s, port = %d\n", serv_ip, serv_port)


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
