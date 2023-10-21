package main

import (
	"context"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type point struct {
	Pid  string  `bson:"hotelId"`
	Plat float64 `bson:"lat"`
	Plon float64 `bson:"lon"`
}

func initializeDatabase(ctx context.Context, url string) *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer client.Close()
	log.Info().Msg("New client successfull...")

	log.Info().Msg("Generating test data...")
	c := client.Database("geo-db").Collection("geo")
	count, err := c.CountDocuments(ctx, bson.M{"hotelId": "1"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &point{"1", 37.7867, -122.4112})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "2"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &point{"2", 37.7854, -122.4005})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "3"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &point{"3", 37.7854, -122.4071})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "4"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &point{"4", 37.7936, -122.3930})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "5"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &point{"5", 37.7831, -122.4181})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "6"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &point{"6", 37.7863, -122.4015})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		hotel_id := strconv.Itoa(i)
		count, err = c.CountDocuments(ctx, bson.M{"hotelId": hotel_id})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		lat := 37.7835 + float64(i)/500.0*3
		lon := -122.41 + float64(i)/500.0*4
		if count == 0 {
			_, err = c.InsertOne(ctx, &point{hotel_id, lat, lon})
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
		}
	}

	// err = c.EnsureIndexKey("hotelId")
	// if err != nil {
	// 	log.Fatal().Msg(err.Error())
	// }

	return client
}
