package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/picop-rd/proxy-controller/app/api/http/client"
	"github.com/picop-rd/proxy-controller/app/entity"
)

const (
	Namespace = "dsb-hr"
	Port      = "9000"
)

var proxyIDs = []string{
	"mongodb-geo",
	"mongodb-profile",
	"mongodb-rate",
	"mongodb-recommendation",
	"mongodb-reservation",
	"mongodb-user",
	"memcached-profile",
	"memcached-rate",
	"memcached-reserve",
}

func main() {
	base := flag.String("url", "http://localhost:8080", "base url for the proxy controller")

	flag.Parse()

	c := client.NewClient(http.DefaultClient, *base)
	if err := registerProxies(c); err != nil {
		log.Fatal(err)
	}
}

func registerProxies(c *client.Client) error {
	pc := client.NewProxy(c)
	for _, id := range proxyIDs {
		proxy := entity.Proxy{
			ProxyID:  id,
			Endpoint: fmt.Sprintf("%s.%s.svc.cluster.local:%s", id, Namespace, Port),
		}
		if err := pc.Register(context.Background(), proxy); err != nil {
			return err
		}
		log.Printf("Registered: ProxyID: %s\n", id)
	}
	return nil
}
