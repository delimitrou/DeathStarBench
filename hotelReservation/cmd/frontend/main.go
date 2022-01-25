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
	"github.com/harlow/go-micro-services/services/frontend"
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

	serv_port, _ := strconv.Atoi(result["FrontendPort"])
	serv_ip := result["FrontendIP"]

	fmt.Printf("frontend ip = %s, port = %d\n upd", serv_ip, serv_port)
	fmt.Println("Waiting...")
	var (
		// port       = flag.Int("port", 5000, "The server port")
		jaegeraddr = flag.String("jaegeraddr", result["jaegerAddress"], "Jaeger address")
		consuladdr = flag.String("consuladdr", result["consulAddress"], "Consul address")
	)
	fmt.Println("Still waiting...")
	flag.Parse()
	fmt.Println("Already waited")
	fmt.Println("Creating tracer...")

	tracer, err := tracing.Init("frontend", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Tracer created: %v", tracer)
	fmt.Println("Creating registry...")

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Registry created: %v", registry)

	srv := &frontend.Server{
		Registry: registry,
		Tracer:   tracer,
		IpAddr:   serv_ip,
		Port:     serv_port,
	}

	fmt.Printf("Preparing server: %v", srv)
	log.Fatal(srv.Run())
}
