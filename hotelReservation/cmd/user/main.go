package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/harlow/go-micro-services/registry"
	"github.com/harlow/go-micro-services/services/user"
	"github.com/harlow/go-micro-services/tracing"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func main() {
	// initializeDatabase()
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]string
	json.Unmarshal([]byte(byteValue), &result)

	serv_port, _ := strconv.Atoi(result["UserPort"])
	serv_ip := ""
	user_mongo_addr := ""
	jaegeraddr := flag.String("jaegeraddr", "", "Jaeger address")
	consuladdr := flag.String("consuladdr", "", "Consul address")

	if result["Orchestrator"] == "k8s"{
		user_mongo_addr = "mongodb-user:"+strings.Split(result["UserMongoAddress"], ":")[1]
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
		user_mongo_addr = result["UserMongoAddress"]
		serv_ip = result["UserIP"]
		*jaegeraddr = result["jaegerAddress"]
		*consuladdr = result["consulAddress"]
	}
	flag.Parse()
	

	mongo_session := initializeDatabase(user_mongo_addr)
	defer mongo_session.Close()

	fmt.Printf("user ip = %s, port = %d\n", serv_ip, serv_port)



	tracer, err := tracing.Init("user", *jaegeraddr)
	if err != nil {
		panic(err)
	}

	registry, err := registry.NewClient(*consuladdr)
	if err != nil {
		panic(err)
	}

	srv := &user.Server{
		Tracer:   tracer,
		// Port:     *port,
		Registry: registry,
		Port:     serv_port,
		IpAddr:	  serv_ip,
		MongoSession: mongo_session,
	}
	log.Fatal(srv.Run())
}
