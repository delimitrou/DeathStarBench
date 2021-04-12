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
	serv_port, _ := strconv.Atoi(result["GeoPort"])
	serv_ip := ""
	geo_mongo_addr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s"{
		geo_mongo_addr = "mongodb-geo:"+strings.Split(result["GeoMongoAddress"], ":")[1]
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
		geo_mongo_addr = result["GeoMongoAddress"]
		serv_ip = result["GeoIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()

	mongo_session := initializeDatabase(geo_mongo_addr)
	defer mongo_session.Close()


	fmt.Printf("geo ip = %s, port = %d\n", serv_ip, serv_port)
	

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
