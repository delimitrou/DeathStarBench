package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/picop-rd/picop-go/contrib/go.mongodb.org/mongo-driver/mongo/picopmongo"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

func initializeDatabase(ctx context.Context, url string) *picopmongo.Client {
	pc := picopmongo.New(options.Client().ApplyURI(url))
	client, err := pc.Connect(ctx)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer client.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := client.Database("user-db").Collection("user")
	for i := 0; i <= 500; i++ {
		suffix := strconv.Itoa(i)
		user_name := "Cornell_" + suffix
		password := ""
		for j := 0; j < 10; j++ {
			password += suffix
		}

		count, err := c.CountDocuments(ctx, bson.M{"username": user_name})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		if count == 0 {
			sum := sha256.Sum256([]byte(password))
			pass := fmt.Sprintf("%x", sum)
			_, err = c.InsertOne(ctx, &User{user_name, pass})
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
		}

	}

	// err = c.EnsureIndexKey("username")
	// if err != nil {
	// 	log.Fatal().Msg(err.Error())
	// }

	return pc

	// count, err := c.Find(&bson.M{"username": "Cornell"}).Count()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if count == 0{
	// 	err = c.Insert(&User{"Cornell", "302eacf716390b1ebb39012b130302efec8a32ac4b8ad0a911112c53b60382b0"})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// count, err = c.Find(&bson.M{"username": "ECE"}).Count()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if count == 0{
	// 	err = c.Insert(&User{"ECE", "a0a44ed8cfc32b7e61befeb99bbff7706808c3fe4dcdf4750a8addb3ffcd4008"})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
}
