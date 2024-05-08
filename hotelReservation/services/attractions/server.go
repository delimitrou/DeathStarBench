package attractions

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/registry"
	pb "github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/services/attractions/proto"
	"github.com/delimitrou/DeathStarBench/tree/master/hotelReservation/tls"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/hailocab/go-geoindex"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

const (
	name             = "srv-attractions"
	maxSearchRadius  = 10
	maxSearchResults = 5
)

// Server implements the attractions service
type Server struct {
	pb.UnimplementedAttractionsServer

	indexH *geoindex.ClusteringIndex
	indexR *geoindex.ClusteringIndex
	indexM *geoindex.ClusteringIndex
	indexC *geoindex.ClusteringIndex
	uuid   string

	Registry    *registry.Client
	Tracer      opentracing.Tracer
	Port        int
	IpAddr      string
	MongoClient *mongo.Client
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.indexH == nil {
		s.indexH = newGeoIndex(s.MongoClient)
	}

	if s.indexR == nil {
		s.indexR = newGeoIndexRest(s.MongoClient)
	}

	if s.indexM == nil {
		s.indexM = newGeoIndexMus(s.MongoClient)
	}

	if s.indexC == nil {
		s.indexC = newGeoIndexCinema(s.MongoClient)
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
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)

	pb.RegisterAttractionsServer(srv, s)

	// listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	err = s.Registry.Register(name, s.uuid, s.IpAddr, s.Port)
	if err != nil {
		return fmt.Errorf("failed register: %v", err)
	}
	log.Info().Msg("Successfully registered in consul")

	return srv.Serve(lis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
	s.Registry.Deregister(s.uuid)
}

// NearbyRest returns all restaurants close to the hotel.
func (s *Server) NearbyRest(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	log.Trace().Msgf("In Attractions NearbyRest")

	mongoSpan, _ := opentracing.StartSpanFromContext(ctx, "mongo_restaurant")
	mongoSpan.SetTag("span.kind", "client")

	c := s.MongoClient.Database("attractions-db").Collection("hotels")

	curr, err := c.Find(context.TODO(), bson.M{"hotelId": req.HotelId})
	if err != nil {
		log.Error().Msgf("Failed get hotels: ", err)
	}
	var hotelReqs []point
	curr.All(context.TODO(), &hotelReqs)

	var hotelReq point

	for _, hotelHelper := range hotelReqs {
		hotelReq = hotelHelper
	}

	var (
		points = s.getNearbyPointsRest(ctx, float64(hotelReq.Plat), float64(hotelReq.Plon))
		res    = &pb.Result{}
	)

	log.Trace().Msgf("restaurants after getNearbyPoints, len = %d", len(points))

	for _, p := range points {
		log.Trace().Msgf("In restaurants Nearby return restaurantId = %s", p.Id())
		res.AttractionIds = append(res.AttractionIds, p.Id())
	}

	return res, nil
}

// NearbyMus returns all museums close to the hotel.
func (s *Server) NearbyMus(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	log.Trace().Msgf("In Attractions NearbyMus")

	mongoSpan, _ := opentracing.StartSpanFromContext(ctx, "mongo_museum")
	mongoSpan.SetTag("span.kind", "client")

	c := s.MongoClient.Database("attractions-db").Collection("hotels")

	curr, err := c.Find(context.TODO(), bson.M{"hotelId": req.HotelId})
	if err != nil {
		log.Error().Msgf("Failed get hotels: ", err)
	}
	var hotelReqs []point
	curr.All(context.TODO(), &hotelReqs)

	var hotelReq point

	for _, hotelHelper := range hotelReqs {
		hotelReq = hotelHelper
	}

	var (
		points = s.getNearbyPointsMus(ctx, float64(hotelReq.Plat), float64(hotelReq.Plon))
		res    = &pb.Result{}
	)

	log.Trace().Msgf("museums after getNearbyPoints, len = %d", len(points))

	for _, p := range points {
		log.Trace().Msgf("In museums Nearby return museumId = %s", p.Id())
		res.AttractionIds = append(res.AttractionIds, p.Id())
	}

	return res, nil
}

// NearbyCinema returns all cinemas close to the hotel.
func (s *Server) NearbyCinema(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	log.Trace().Msgf("In Attractions NearbyCinema")

	mongoSpan, _ := opentracing.StartSpanFromContext(ctx, "mongo_cinema")
	mongoSpan.SetTag("span.kind", "client")

	c := s.MongoClient.Database("attractions-db").Collection("hotels")

	curr, err := c.Find(context.TODO(), bson.M{"hotelId": req.HotelId})
	if err != nil {
		log.Error().Msgf("Failed get hotels: ", err)
	}
	var hotelReqs []point
	curr.All(context.TODO(), &hotelReqs)

	var hotelReq point

	for _, hotelHelper := range hotelReqs {
		hotelReq = hotelHelper
	}

	var (
		points = s.getNearbyPointsCinema(ctx, float64(hotelReq.Plat), float64(hotelReq.Plon))
		res    = &pb.Result{}
	)

	log.Trace().Msgf("cinemas after getNearbyPoints, len = %d", len(points))

	for _, p := range points {
		log.Trace().Msgf("In cinemas Nearby return cinemaId = %s", p.Id())
		res.AttractionIds = append(res.AttractionIds, p.Id())
	}

	return res, nil
}

func (s *Server) getNearbyPointsHotel(ctx context.Context, lat, lon float64) []geoindex.Point {
	log.Trace().Msgf("In geo getNearbyPoints, lat = %f, lon = %f", lat, lon)

	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.indexH.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
}

func (s *Server) getNearbyPointsRest(ctx context.Context, lat, lon float64) []geoindex.Point {
	log.Trace().Msgf("In geo getNearbyPointsRest, lat = %f, lon = %f", lat, lon)

	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.indexR.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
}

func (s *Server) getNearbyPointsMus(ctx context.Context, lat, lon float64) []geoindex.Point {
	log.Trace().Msgf("In geo getNearbyPointsMus, lat = %f, lon = %f", lat, lon)

	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.indexM.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
}

func (s *Server) getNearbyPointsCinema(ctx context.Context, lat, lon float64) []geoindex.Point {
	log.Trace().Msgf("In geo getNearbyPointsCinema, lat = %f, lon = %f", lat, lon)

	center := &geoindex.GeoPoint{
		Pid:  "",
		Plat: lat,
		Plon: lon,
	}

	return s.indexC.KNearest(
		center,
		maxSearchResults,
		geoindex.Km(maxSearchRadius), func(p geoindex.Point) bool {
			return true
		},
	)
}

// newGeoIndex returns a geo index with points loaded
func newGeoIndex(client *mongo.Client) *geoindex.ClusteringIndex {
	log.Trace().Msg("new geo newGeoIndex")

	collection := client.Database("attractions-db").Collection("hotels")
	curr, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Error().Msgf("Failed get hotels data: ", err)
	}

	var points []*point
	curr.All(context.TODO(), &points)
	if err != nil {
		log.Error().Msgf("Failed get hotels data: ", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

// newGeoIndexRest returns a geo index with points loaded
func newGeoIndexRest(client *mongo.Client) *geoindex.ClusteringIndex {
	log.Trace().Msg("new geo newGeoIndexRest")

	collection := client.Database("attractions-db").Collection("restaurants")
	curr, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Error().Msgf("Failed get restaurant data: ", err)
	}

	var points []*Restaurant
	curr.All(context.TODO(), &points)
	if err != nil {
		log.Error().Msgf("Failed get restaurant data: ", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

// newGeoIndexMus returns a geo index with points loaded
func newGeoIndexMus(client *mongo.Client) *geoindex.ClusteringIndex {
	log.Trace().Msg("new geo newGeoIndexMus")

	collection := client.Database("attractions-db").Collection("museums")
	curr, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Error().Msgf("Failed get restaurant data: ", err)
	}

	var points []*Museum
	curr.All(context.TODO(), &points)
	if err != nil {
		log.Error().Msgf("Failed get restaurant data: ", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

// newGeoIndexCinema returns a geo index with points loaded
func newGeoIndexCinema(client *mongo.Client) *geoindex.ClusteringIndex {
	log.Trace().Msg("new geo newGeoIndexCinema")

	collection := client.Database("attractions-db").Collection("cinemas")
	curr, err := collection.Find(context.TODO(), bson.D{})
	if err != nil {
		log.Error().Msgf("Failed get cinema data: ", err)
	}

	var points []*Cinema
	curr.All(context.TODO(), &points)
	if err != nil {
		log.Error().Msgf("Failed get cinema data: ", err)
	}

	// add points to index
	index := geoindex.NewClusteringIndex()
	for _, point := range points {
		index.Add(point)
	}

	return index
}

type point struct {
	Pid  string  `bson:"hotelId"`
	Plat float64 `bson:"lat"`
	Plon float64 `bson:"lon"`
}

// Implement Point interface
func (p *point) Lat() float64 { return p.Plat }
func (p *point) Lon() float64 { return p.Plon }
func (p *point) Id() string   { return p.Pid }

type Restaurant struct {
	RestaurantId   string  `bson:"restaurantId"`
	RLat           float64 `bson:"lat"`
	RLon           float64 `bson:"lon"`
	RestaurantName string  `bson:"restaurantName"`
	Rating         float32 `bson:"rating"`
	Type           string  `bson:"type"`
}

// Implement Restaurant interface
func (r *Restaurant) Lat() float64 { return r.RLat }
func (r *Restaurant) Lon() float64 { return r.RLon }
func (r *Restaurant) Id() string   { return r.RestaurantId }

type Museum struct {
	MuseumId   string  `bson:"museumId"`
	MLat       float64 `bson:"lat"`
	MLon       float64 `bson:"lon"`
	MuseumName string  `bson:"museumName"`
	Type       string  `bson:"type"`
}

// Implement Museum interface
func (m *Museum) Lat() float64 { return m.MLat }
func (m *Museum) Lon() float64 { return m.MLon }
func (m *Museum) Id() string   { return m.MuseumId }

type Cinema struct {
	CinemaId   string  `bson:"cinemaId"`
	CLat       float64 `bson:"lat"`
	CLon       float64 `bson:"lon"`
	CinemaName string  `bson:"cinemaName"`
	Type       string  `bson:"type"`
}

// Implement Cinema interface
func (c *Cinema) Lat() float64 { return c.CLat }
func (c *Cinema) Lon() float64 { return c.CLon }
func (c *Cinema) Id() string   { return c.CinemaId }
