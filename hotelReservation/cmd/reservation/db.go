package main

import (
	"context"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Reservation struct {
	HotelId      string `bson:"hotelId"`
	CustomerName string `bson:"customerName"`
	InDate       string `bson:"inDate"`
	OutDate      string `bson:"outDate"`
	Number       int    `bson:"number"`
}

type Number struct {
	HotelId string `bson:"hotelId"`
	Number  int    `bson:"numberOfRoom"`
}

func initializeDatabase(ctx context.Context, url string) *mongo.Client {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer client.Close()
	log.Info().Msg("New session successfull...")

	c := client.Database("reservation-db").Collection("reservation")
	count, err := c.CountDocuments(ctx, bson.M{"hotelId": "4"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Reservation{"4", "Alice", "2015-04-09", "2015-04-10", 1})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	c = client.Database("reservation-db").Collection("number")
	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "1"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Number{"1", 200})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "2"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Number{"2", 200})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "3"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Number{"3", 200})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "4"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Number{"4", 200})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "5"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Number{"5", 200})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.CountDocuments(ctx, bson.M{"hotelId": "6"})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		_, err = c.InsertOne(ctx, &Number{"6", 200})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	for i := 7; i <= 80; i++ {
		hotel_id := strconv.Itoa(i)
		count, err = c.CountDocuments(ctx, bson.M{"hotelId": hotel_id})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		room_num := 200
		if i%3 == 1 {
			room_num = 300
		} else if i%3 == 2 {
			room_num = 250
		}
		if count == 0 {
			_, err = c.InsertOne(ctx, &Number{hotel_id, room_num})
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
