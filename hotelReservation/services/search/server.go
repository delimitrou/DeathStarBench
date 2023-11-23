package search

import (
	// "encoding/json"
	"fmt"
	// F"io"
	"context"
	"net"

	"github.com/rs/zerolog/log"

	// "os"
	"time"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/dialer"
	geo "github.com/harlow/go-micro-services/services/geo/proto"
	rate "github.com/harlow/go-micro-services/services/rate/proto"
	pb "github.com/harlow/go-micro-services/services/search/proto"
	"github.com/harlow/go-micro-services/tls"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/picop-rd/picop-go/contrib/google.golang.org/grpc/picopgrpc"
	"github.com/picop-rd/picop-go/propagation"
	picopnet "github.com/picop-rd/picop-go/protocol/net"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const name = "search"

// Server implments the search service
type Server struct {
	geoClient  *picopgrpc.Client
	rateClient *picopgrpc.Client

	Tracer     opentracing.Tracer
	Port       int
	IpAddr     string
	KnativeDns string
	uuid       string
}

// Run starts the server
func (s *Server) Run() error {
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
	pb.RegisterSearchServer(srv, s)

	// init grpc clients
	if err := s.initGeoClient("geo:8083"); err != nil {
		return err
	}
	if err := s.initRateClient("rate:8084"); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	blis := picopnet.NewListener(lis)

	// register with consul
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

func (s *Server) initGeoClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.geoClient = conn
	return nil
}

func (s *Server) initRateClient(name string) error {
	conn, err := s.getGprcConn(name)
	if err != nil {
		return fmt.Errorf("dialer error: %v", err)
	}
	s.rateClient = conn
	return nil
}

func (s *Server) getGprcConn(name string) (*picopgrpc.Client, error) {
	if s.KnativeDns != "" {
		return dialer.Dial(
			fmt.Sprintf("%s.%s", name, s.KnativeDns),
			dialer.WithTracer(s.Tracer))
	} else {
		return dialer.Dial(
			name,
			dialer.WithTracer(s.Tracer),
		)
	}
}

// Nearby returns ids of nearby hotels ordered by ranking algo
func (s *Server) Nearby(ctx context.Context, req *pb.NearbyRequest) (*pb.SearchResult, error) {
	// find nearby hotels
	log.Trace().Msg("in Search Nearby")

	log.Trace().Msgf("nearby lat = %f", req.Lat)
	log.Trace().Msgf("nearby lon = %f", req.Lon)

	geoConn, err := s.geoClient.Connect(ctx)
	if err != nil {
		return nil, err
	}
	geoClient := geo.NewGeoClient(geoConn)
	nearby, err := geoClient.Nearby(ctx, &geo.Request{
		Lat: req.Lat,
		Lon: req.Lon,
	})
	if err != nil {
		return nil, err
	}

	for _, hid := range nearby.HotelIds {
		log.Trace().Msgf("get Nearby hotelId = %s", hid)
	}

	// find rates for hotels
	rateConn, err := s.rateClient.Connect(ctx)
	if err != nil {
		return nil, err
	}
	rateClient := rate.NewRateClient(rateConn)
	rates, err := rateClient.GetRates(ctx, &rate.Request{
		HotelIds: nearby.HotelIds,
		InDate:   req.InDate,
		OutDate:  req.OutDate,
	})
	if err != nil {
		return nil, err
	}

	// TODO(hw): add simple ranking algo to order hotel ids:
	// * geo distance
	// * price (best discount?)
	// * reviews

	// build the response
	res := new(pb.SearchResult)
	for _, ratePlan := range rates.RatePlans {
		log.Trace().Msgf("get RatePlan HotelId = %s, Code = %s", ratePlan.HotelId, ratePlan.Code)
		res.HotelIds = append(res.HotelIds, ratePlan.HotelId)
	}
	return res, nil
}
