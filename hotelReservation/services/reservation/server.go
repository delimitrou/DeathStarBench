package reservation

import (
	// "encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/harlow/go-micro-services/registry"
	pb "github.com/harlow/go-micro-services/services/reservation/proto"
	"github.com/harlow/go-micro-services/tls"
	"github.com/opentracing/opentracing-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	// "io/ioutil"
	"net"
	// "os"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/rs/zerolog/log"

	// "strings"
	"strconv"
)

const name = "srv-reservation"

// Server implements the user service
type Server struct {
	Tracer       opentracing.Tracer
	Port         int
	IpAddr       string
	MongoSession *mgo.Session
	Registry     *registry.Client
	MemcClient   *memcache.Client
	uuid         string
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
		grpc.UnaryInterceptor(
			otgrpc.OpenTracingServerInterceptor(s.Tracer),
		),
	}

	if tlsopt := tls.GetServerOpt(); tlsopt != nil {
		opts = append(opts, tlsopt)
	}

	srv := grpc.NewServer(opts...)

	pb.RegisterReservationServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}

	// register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := ioutil.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

	log.Trace().Msgf("In reservation s.IpAddr = %s, port = %d", s.IpAddr, s.Port)

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

// MakeReservation makes a reservation based on given information
func (s *Server) MakeReservation(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	res.HotelId = make([]string, 0)

	// session, err := mgo.Dial("mongodb-reservation")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	session := s.MongoSession.Copy()
	defer session.Close()

	c := session.DB("reservation-db").C("reservation")
	c1 := session.DB("reservation-db").C("number")

	inDate, _ := time.Parse(
		time.RFC3339,
		req.InDate+"T12:00:00+00:00")

	outDate, _ := time.Parse(
		time.RFC3339,
		req.OutDate+"T12:00:00+00:00")
	hotelId := req.HotelId[0]

	indate := inDate.String()[0:10]

	memc_date_num_map := make(map[string]int)

	for inDate.Before(outDate) {
		// check reservations
		count := 0
		inDate = inDate.AddDate(0, 0, 1)
		outdate := inDate.String()[0:10]

		// first check memc
		memc_key := hotelId + "_" + inDate.String()[0:10] + "_" + outdate
		item, err := s.MemcClient.Get(memc_key)
		if err == nil {
			// memcached hit
			count, _ = strconv.Atoi(string(item.Value))
			log.Trace().Msgf("memcached hit %s = %d", memc_key, count)
			memc_date_num_map[memc_key] = count + int(req.RoomNumber)

		} else if err == memcache.ErrCacheMiss {
			// memcached miss
			log.Trace().Msgf("memcached miss")
			reserve := make([]reservation, 0)
			err := c.Find(&bson.M{"hotelId": hotelId, "inDate": indate, "outDate": outdate}).All(&reserve)
			if err != nil {
				log.Panic().Msgf("Tried to find hotelId [%v] from date [%v] to date [%v], but got error", hotelId, indate, outdate, err.Error())
			}

			for _, r := range reserve {
				count += r.Number
			}

			memc_date_num_map[memc_key] = count + int(req.RoomNumber)

		} else {
			log.Panic().Msgf("Tried to get memc_key [%v], but got memmcached error = %s", memc_key, err)
		}

		// check capacity
		// check memc capacity
		memc_cap_key := hotelId + "_cap"
		item, err = s.MemcClient.Get(memc_cap_key)
		hotel_cap := 0
		if err == nil {
			// memcached hit
			hotel_cap, _ = strconv.Atoi(string(item.Value))
			log.Trace().Msgf("memcached hit %s = %d", memc_cap_key, hotel_cap)
		} else if err == memcache.ErrCacheMiss {
			// memcached miss
			var num number
			err = c1.Find(&bson.M{"hotelId": hotelId}).One(&num)
			if err != nil {
				log.Panic().Msgf("Tried to find hotelId [%v], but got error", hotelId, err.Error())
			}
			hotel_cap = int(num.Number)

			// write to memcache
			s.MemcClient.Set(&memcache.Item{Key: memc_cap_key, Value: []byte(strconv.Itoa(hotel_cap))})
		} else {
			log.Panic().Msgf("Tried to get memc_cap_key [%v], but got memmcached error = %s", memc_cap_key, err)
		}

		if count+int(req.RoomNumber) > hotel_cap {
			return res, nil
		}
		indate = outdate
	}

	// only update reservation number cache after check succeeds
	for key, val := range memc_date_num_map {
		s.MemcClient.Set(&memcache.Item{Key: key, Value: []byte(strconv.Itoa(val))})
	}

	inDate, _ = time.Parse(
		time.RFC3339,
		req.InDate+"T12:00:00+00:00")

	indate = inDate.String()[0:10]

	for inDate.Before(outDate) {
		inDate = inDate.AddDate(0, 0, 1)
		outdate := inDate.String()[0:10]
		err := c.Insert(&reservation{
			HotelId:      hotelId,
			CustomerName: req.CustomerName,
			InDate:       indate,
			OutDate:      outdate,
			Number:       int(req.RoomNumber)})
		if err != nil {
			log.Panic().Msgf("Tried to insert hotel [hotelId %v], but got error", hotelId, err.Error())
		}
		indate = outdate
	}

	res.HotelId = append(res.HotelId, hotelId)

	return res, nil
}

// CheckAvailability checks if given information is available
func (s *Server) CheckAvailability(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)
	res.HotelId = make([]string, 0)

	// session, err := mgo.Dial("mongodb-reservation")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	session := s.MongoSession.Copy()
	defer session.Close()

	c := session.DB("reservation-db").C("reservation")
	c1 := session.DB("reservation-db").C("number")

	for _, hotelId := range req.HotelId {
		log.Trace().Msgf("reservation check hotel %s", hotelId)
		inDate, _ := time.Parse(
			time.RFC3339,
			req.InDate+"T12:00:00+00:00")

		outDate, _ := time.Parse(
			time.RFC3339,
			req.OutDate+"T12:00:00+00:00")

		indate := inDate.String()[0:10]

		for inDate.Before(outDate) {
			// check reservations
			count := 0
			inDate = inDate.AddDate(0, 0, 1)
			log.Trace().Msgf("reservation check date %s", inDate.String()[0:10])
			outdate := inDate.String()[0:10]

			// first check memc
			memc_key := hotelId + "_" + inDate.String()[0:10] + "_" + outdate
			item, err := s.MemcClient.Get(memc_key)

			if err == nil {
				// memcached hit
				count, _ = strconv.Atoi(string(item.Value))
				log.Trace().Msgf("memcached hit %s = %d", memc_key, count)
			} else if err == memcache.ErrCacheMiss {
				// memcached miss
				reserve := make([]reservation, 0)
				err := c.Find(&bson.M{"hotelId": hotelId, "inDate": indate, "outDate": outdate}).All(&reserve)
				if err != nil {
					log.Panic().Msgf("Tried to find hotelId [%v] from date [%v] to date [%v], but got error", hotelId, indate, outdate, err.Error())
				}
				for _, r := range reserve {
					log.Trace().Msgf("reservation check reservation number = %d", hotelId)
					count += r.Number
				}

				// update memcached
				s.MemcClient.Set(&memcache.Item{Key: memc_key, Value: []byte(strconv.Itoa(count))})
			} else {
				log.Panic().Msgf("Tried to get memc_key [%v], but got memmcached error = %s", memc_key, err)

			}

			// check capacity
			// check memc capacity
			memc_cap_key := hotelId + "_cap"
			item, err = s.MemcClient.Get(memc_cap_key)
			hotel_cap := 0

			if err == nil {
				// memcached hit
				hotel_cap, _ = strconv.Atoi(string(item.Value))
				log.Trace().Msgf("memcached hit %s = %d", memc_cap_key, hotel_cap)
			} else if err == memcache.ErrCacheMiss {
				var num number
				err = c1.Find(&bson.M{"hotelId": hotelId}).One(&num)
				if err != nil {
					log.Panic().Msgf("Tried to find hotelId [%v], but got error", hotelId, err.Error())
				}
				hotel_cap = int(num.Number)
				// update memcached
				s.MemcClient.Set(&memcache.Item{Key: memc_cap_key, Value: []byte(strconv.Itoa(hotel_cap))})
			} else {
				log.Panic().Msgf("Tried to get memc_key [%v], but got memmcached error = %s", memc_cap_key, err)
			}

			if count+int(req.RoomNumber) > hotel_cap {
				break
			}
			indate = outdate

			if inDate.Equal(outDate) {
				res.HotelId = append(res.HotelId, hotelId)
			}
		}
	}

	return res, nil
}

type reservation struct {
	HotelId      string `bson:"hotelId"`
	CustomerName string `bson:"customerName"`
	InDate       string `bson:"inDate"`
	OutDate      string `bson:"outDate"`
	Number       int    `bson:"number"`
}

type number struct {
	HotelId string `bson:"hotelId"`
	Number  int    `bson:"numberOfRoom"`
}
