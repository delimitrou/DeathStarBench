package rate

import (
	"context"
	"encoding/json"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	// "io"
	"net"
	// "os"
	"sort"
	"sync"
	"time"

	"github.com/picop-rd/picop-go/contrib/go.mongodb.org/mongo-driver/mongo/picopmongo"
	"github.com/rs/zerolog/log"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	pb "github.com/harlow/go-micro-services/services/rate/proto"
	"github.com/harlow/go-micro-services/tls"
	"github.com/opentracing/opentracing-go"
	picopmc "github.com/picop-rd/picop-go/contrib/github.com/bradfitz/gomemcache/picopgomemcache"
	"github.com/picop-rd/picop-go/contrib/google.golang.org/grpc/picopgrpc"
	"github.com/picop-rd/picop-go/propagation"
	picopnet "github.com/picop-rd/picop-go/protocol/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"strings"

	"github.com/bradfitz/gomemcache/memcache"
)

const name = "srv-rate"

// Server implements the rate service
type Server struct {
	Tracer      opentracing.Tracer
	Port        int
	IpAddr      string
	MongoClient *picopmongo.Client
	MemcClient  *picopmc.Client
	uuid        string
}

// Run starts the server
func (s *Server) Run() error {
	opentracing.SetGlobalTracer(s.Tracer)

	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	s.uuid = uuid.New().String()

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: 120 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.ChainUnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
			picopgrpc.UnaryServerInterceptor(propagation.EnvID{}),
		),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)

	pb.RegisterRateServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	blis := picopnet.NewListener(lis)

	// register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := io.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

	return srv.Serve(blis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// GetRates gets rates for hotels for specific date range.
func (s *Server) GetRates(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	// session, err := mgo.Dial("mongodb-rate")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	ratePlans := make(RatePlans, 0)

	hotelIds := []string{}
	rateMap := make(map[string]struct{})
	for _, hotelID := range req.HotelIds {
		hotelIds = append(hotelIds, hotelID)
		rateMap[hotelID] = struct{}{}
	}
	mclient := s.MemcClient.Connect(ctx)
	// first check memcached(get-multi)
	memSpan, _ := opentracing.StartSpanFromContext(ctx, "memcached_get_multi_rate")
	memSpan.SetTag("span.kind", "client")
	resMap, err := mclient.GetMulti(ctx, hotelIds)
	memSpan.Finish()
	var wg sync.WaitGroup
	var mutex sync.Mutex
	if err != nil && err != memcache.ErrCacheMiss {
		log.Panic().Msgf("Memmcached error while trying to get hotel [id: %v]= %s", hotelIds, err)
	} else {
		for hotelId, item := range resMap {
			rateStrs := strings.Split(string(item.Value), "\n")
			log.Trace().Msgf("memc hit, hotelId = %s,rate strings: %v", hotelId, rateStrs)

			for _, rateStr := range rateStrs {
				if len(rateStr) != 0 {
					rateP := new(pb.RatePlan)
					json.Unmarshal([]byte(rateStr), rateP)
					ratePlans = append(ratePlans, rateP)
				}
			}
			delete(rateMap, hotelId)
		}
		client, err := s.MongoClient.Connect(ctx)
		if err != nil {
			log.Panic().Msgf("Got error while connecting to mongo: %v", err)
		}
		wg.Add(len(rateMap))
		for hotelId := range rateMap {
			go func(id string) {
				log.Trace().Msgf("memc miss, hotelId = %s", id)
				log.Trace().Msg("memcached miss, set up mongo connection")

				// memcached miss, set up mongo connection
				c := client.Database("rate-db").Collection("inventory")
				memcStr := ""
				tmpRatePlans := make(RatePlans, 0)
				mongoSpan, _ := opentracing.StartSpanFromContext(ctx, "mongo_rate")
				mongoSpan.SetTag("span.kind", "client")
				cur, err := c.Find(ctx, &bson.M{"hotelId": id})
				if err != nil {
					log.Panic().Msgf("Tried to find hotelId [%v], but got error", id, err.Error())
				}
				err = cur.All(ctx, &tmpRatePlans)
				mongoSpan.Finish()
				if err != nil {
					log.Panic().Msgf("Tried to find hotelId [%v], but got error", id, err.Error())
				} else {
					for _, r := range tmpRatePlans {
						mutex.Lock()
						ratePlans = append(ratePlans, r)
						mutex.Unlock()
						rateJson, err := json.Marshal(r)
						if err != nil {
							log.Error().Msgf("Failed to marshal plan [Code: %v] with error: %s", r.Code, err)
						}
						memcStr = memcStr + string(rateJson) + "\n"
					}
				}
				go mclient.Set(ctx, &memcache.Item{Key: id, Value: []byte(memcStr)})

				defer wg.Done()
			}(hotelId)
		}
	}
	wg.Wait()

	sort.Sort(ratePlans)
	res.RatePlans = ratePlans

	return res, nil
}

type RatePlans []*pb.RatePlan

func (r RatePlans) Len() int {
	return len(r)
}

func (r RatePlans) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r RatePlans) Less(i, j int) bool {
	return r[i].RoomType.TotalRate > r[j].RoomType.TotalRate
}
