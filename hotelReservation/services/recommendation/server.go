package recommendation

import (
	// "encoding/json"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/hailocab/go-geoindex"
	"github.com/harlow/go-micro-services/registry"
	pb "github.com/harlow/go-micro-services/services/recommendation/proto"
	"github.com/harlow/go-micro-services/tls"
	"github.com/opentracing/opentracing-go"
	"github.com/picop-rd/picop-go/contrib/go.mongodb.org/mongo-driver/mongo/picopmongo"
	"github.com/picop-rd/picop-go/contrib/google.golang.org/grpc/picopgrpc"
	"github.com/picop-rd/picop-go/propagation"
	picopnet "github.com/picop-rd/picop-go/protocol/net"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// "io"
	"math"
	"net"

	// "os"
	"time"
	// "strings"
)

const name = "srv-recommendation"

// Server implements the recommendation service
type Server struct {
	hotels      map[string]Hotel
	Tracer      opentracing.Tracer
	Port        int
	IpAddr      string
	MongoClient *picopmongo.Client
	Registry    *registry.Client
	uuid        string
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.hotels == nil {
		ctx := context.Background()
		mc, err := s.MongoClient.Connect(ctx)
		if err != nil {
			return fmt.Errorf("Failed connect to mongo: ", err)
		}
		s.hotels = loadRecommendations(ctx, mc)
	}

	s.uuid = uuid.New().String()

	opts := []grpc.ServerOption{
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Timeout: 120 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
		),
		grpc.UnaryInterceptor(
			picopgrpc.UnaryServerInterceptor(propagation.EnvID{}),
		),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)

	pb.RegisterRecommendationServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	blis := picopnet.NewListener(lis)

	// // register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := io.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

	err = s.Registry.Register(name, s.uuid, s.IpAddr, s.Port)
	if err != nil {
		return fmt.Errorf("failed register: %v", err)
	}
	log.Info().Msg("Successfully registered in consul")

	return srv.Serve(blis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
	s.Registry.Deregister(s.uuid)
}

// GiveRecommendation returns recommendations within a given requirement.
func (s *Server) GetRecommendations(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	log.Trace().Msgf("GetRecommendations")
	require := req.Require
	if require == "dis" {
		p1 := &geoindex.GeoPoint{
			Pid:  "",
			Plat: req.Lat,
			Plon: req.Lon,
		}
		min := math.MaxFloat64
		for _, hotel := range s.hotels {
			tmp := float64(geoindex.Distance(p1, &geoindex.GeoPoint{
				Pid:  "",
				Plat: hotel.HLat,
				Plon: hotel.HLon,
			})) / 1000
			if tmp < min {
				min = tmp
			}
		}
		for _, hotel := range s.hotels {
			tmp := float64(geoindex.Distance(p1, &geoindex.GeoPoint{
				Pid:  "",
				Plat: hotel.HLat,
				Plon: hotel.HLon,
			})) / 1000
			if tmp == min {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	} else if require == "rate" {
		max := 0.0
		for _, hotel := range s.hotels {
			if hotel.HRate > max {
				max = hotel.HRate
			}
		}
		for _, hotel := range s.hotels {
			if hotel.HRate == max {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	} else if require == "price" {
		min := math.MaxFloat64
		for _, hotel := range s.hotels {
			if hotel.HPrice < min {
				min = hotel.HPrice
			}
		}
		for _, hotel := range s.hotels {
			if hotel.HPrice == min {
				res.HotelIds = append(res.HotelIds, hotel.HId)
			}
		}
	} else {
		log.Warn().Msgf("Wrong require parameter: %v", require)
	}

	return res, nil
}

// loadRecommendations loads hotel recommendations from mongodb.
func loadRecommendations(ctx context.Context, client *mongo.Client) map[string]Hotel {
	// session, err := mgo.Dial("mongodb-recommendation")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	c := client.Database("recommendation-db").Collection("recommendation")

	// unmarshal json profiles
	var hotels []Hotel
	cur, err := c.Find(ctx, bson.M{})
	if err != nil {
		log.Error().Msgf("Failed get hotels data: ", err)
	}
	err = cur.All(ctx, &hotels)
	if err != nil {
		log.Error().Msgf("Failed get hotels data: ", err)
	}

	profiles := make(map[string]Hotel)
	for _, hotel := range hotels {
		profiles[hotel.HId] = hotel
	}

	return profiles
}

type Hotel struct {
	ID     primitive.ObjectID `bson:"_id"`
	HId    string             `bson:"hotelId"`
	HLat   float64            `bson:"lat"`
	HLon   float64            `bson:"lon"`
	HRate  float64            `bson:"rate"`
	HPrice float64            `bson:"price"`
}
