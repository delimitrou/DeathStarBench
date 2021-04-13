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
	"github.com/harlow/go-micro-services/services/search"
	"github.com/harlow/go-micro-services/tracing"
	"strconv"
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
	serv_ip := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	serv_port, _ := strconv.Atoi(result["SearchPort"])
	if result["Orchestrator"] == "k8s"{
		addrs, _ := net.InterfaceAddrs()
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					serv_ip = ipnet.IP.String()

				}
			}
		}
		*jaegeraddr =  "jaeger:"+strings.Split(result["jaegerAddress"], ":")[1]
		*consuladdr = "consul:" + strings.Split(result["consulAddress"], ":")[1]

	} else {
		serv_ip	= result["SearchIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]

	}
	flag.Parse()



	fmt.Printf("search ip = %s, port = %d\n", serv_ip, serv_port)

	tracer, err := tracing.Init("search", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &search.Server{
		Tracer:   tracer,
		// Port:     *port,
		Port:     serv_port,
		IpAddr:	  serv_ip,
		Registry: registry,
	}
	log.Fatal(srv.Run())
}
