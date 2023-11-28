package user

import (
	"crypto/sha256"
	// "encoding/json"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	pb "github.com/harlow/go-micro-services/services/user/proto"
	"github.com/harlow/go-micro-services/tls"
	"github.com/opentracing/opentracing-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	// "io"
	"net"

	"github.com/rs/zerolog/log"

	// "os"
	"time"
)

const name = "user"

// Server implements the user service
type Server struct {
	users map[string]string

	Tracer      opentracing.Tracer
	Port        int
	IpAddr      string
	MongoClient *mongo.Client
	uuid        string
}

// Run starts the server
func (s *Server) Run() error {
	if s.Port == 0 {
		return fmt.Errorf("server port must be set")
	}

	if s.users == nil {
		s.users = loadUsers(s.MongoClient)
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

	pb.RegisterUserServer(srv, s)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}

	// // register the service
	// jsonFile, err := os.Open("config.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// defer jsonFile.Close()

	// byteValue, _ := io.ReadAll(jsonFile)

	// var result map[string]string
	// json.Unmarshal([]byte(byteValue), &result)

	return srv.Serve(lis)
}

// Shutdown cleans up any processes
func (s *Server) Shutdown() {
}

// CheckUser returns whether the username and password are correct.
func (s *Server) CheckUser(ctx context.Context, req *pb.Request) (*pb.Result, error) {
	res := new(pb.Result)

	log.Trace().Msg("CheckUser")

	sum := sha256.Sum256([]byte(req.Password))
	pass := fmt.Sprintf("%x", sum)

	// session, err := mgo.Dial("mongodb-user")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()

	// c := session.DB("user-db").C("user")

	// user := User{}
	// err = c.Find(bson.M{"username": req.Username}).One(&user)
	// if err != nil {
	// 	panic(err)
	// }
	res.Correct = false
	if true_pass, found := s.users[req.Username]; found {
		res.Correct = pass == true_pass
	}

	// res.Correct = user.Password == pass

	log.Trace().Msgf("CheckUser %d", res.Correct)

	return res, nil
}

// loadUsers loads hotel users from mongodb.
func loadUsers(client *mongo.Client) map[string]string {
	ctx := context.Background()
	// session, err := mgo.Dial("mongodb-user")
	// if err != nil {
	// 	panic(err)
	// }
	// defer session.Close()
	c := client.Database("user-db").Collection("user")

	// unmarshal json profiles
	var users []User
	cur, err := c.Find(ctx, bson.M{})
	if err != nil {
		log.Error().Msgf("Failed get users data: ", err)
	}
	err = cur.All(ctx, &users)
	if err != nil {
		log.Error().Msgf("Failed get users data: ", err)
	}

	res := make(map[string]string)
	for _, user := range users {
		res[user.Username] = user.Password
	}

	log.Trace().Msg("Done load users")

	return res
}

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}
