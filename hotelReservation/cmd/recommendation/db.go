package main

import (
	"context"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Hotel struct {
	HId    string  `bson:"hotelId"`
	HLat   float64 `bson:"lat"`
	HLon   float64 `bson:"lon"`
	HRate  float64 `bson:"rate"`
	HPrice float64 `bson:"price"`
}

func initializeDatabase(ctx context.Context, url string) *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer session.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := client.Database("recommendation-db").Collection("recommendation")
	count, err := c.CountDocuments(ctx, bson.M{"hotelId": "1"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Hotel{"1", 37.7867, -122.4112, 109.00, 150.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "2"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Hotel{"2", 37.7854, -122.4005, 139.00, 120.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "3"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Hotel{"3", 37.7834, -122.4071, 109.00, 190.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "4"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Hotel{"4", 37.7936, -122.3930, 129.00, 160.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "5"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Hotel{"5", 37.7831, -122.4181, 119.00, 140.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "6"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Hotel{"6", 37.7863, -122.4015, 149.00, 200.00})
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

		count, err = c.CountDocuments(ctx, bson.M{"hotelId": hotel_id})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		rate := 135.00
		rate_inc := 179.00
		if i%3 == 0 {
			if i%5 == 0 {
				rate = 109.00
				rate_inc = 123.17
			} else if i%5 == 1 {
				rate = 120.00
				rate_inc = 140.00
			} else if i%5 == 2 {
				rate = 124.00
				rate_inc = 144.00
			} else if i%5 == 3 {
				rate = 132.00
				rate_inc = 158.00
			} else if i%5 == 4 {
				rate = 232.00
				rate_inc = 258.00
			}
		}

		if count == 0 {
			_, err = c.InsertOne(ctx, &Hotel{hotel_id, lat, lon, rate, rate_inc})
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
