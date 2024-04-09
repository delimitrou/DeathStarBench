package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Restaurant struct {
	RestaurantId   string  `bson:"restaurantId"`
	RLat           float64 `bson:"lat"`
	RLon           float64 `bson:"lon"`
	RestaurantName string  `bson:"restaurantName"`
	Rating         float32 `bson:"rating"`
	Type           string  `bson:"type"`
}

type Museum struct {
	MuseumId   string  `bson:"museumId"`
	MLat       float64 `bson:"lat"`
	MLon       float64 `bson:"lon"`
	MuseumName string  `bson:"museumName"`
	Type       string  `bson:"type"`
}

type point struct {
	Pid  string  `bson:"hotelId"`
	Plat float64 `bson:"lat"`
	Plon float64 `bson:"lon"`
}

func initializeDatabase(url string) (*mongo.Client, func()) {

	newPoints := []interface{}{
		point{"1", 37.7867, -122.4112},
		point{"2", 37.7854, -122.4005},
		point{"3", 37.7854, -122.4071},
		point{"4", 37.7936, -122.3930},
		point{"5", 37.7831, -122.4181},
		point{"6", 37.7863, -122.4015},
	}

	newRestaurants := []interface{}{
		&Restaurant{"1", 37.7867, -122.4112, "R1", 3.5, "fusion"},
		&Restaurant{"2", 37.7857, -122.4012, "R2", 3.9, "italian"},
		&Restaurant{"3", 37.7847, -122.3912, "R3", 4.5, "sushi"},
		&Restaurant{"4", 37.7862, -122.4212, "R4", 3.2, "sushi"},
		&Restaurant{"5", 37.7839, -122.4052, "R5", 4.9, "fusion"},
		&Restaurant{"6", 37.7831, -122.3812, "R6", 4.1, "american"},
	}

	newMuseums := []interface{}{
		&Museum{"1", 35.7867, -122.4112, "M1", "history"},
		&Museum{"2", 36.7867, -122.5112, "M2", "history"},
		&Museum{"3", 38.7867, -122.4612, "M3", "nature"},
		&Museum{"4", 37.7867, -122.4912, "M4", "nature"},
		&Museum{"5", 36.9867, -122.4212, "M5", "nature"},
		&Museum{"6", 37.3867, -122.5012, "M6", "technology"},
	}

	uri := fmt.Sprintf("mongodb://%s", url)
	log.Info().Msgf("Attempting connection to %v", uri)

	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msg("Successfully connected to MongoDB")

	collectionH := client.Database("attractions-db").Collection("hotels")
	_, err = collectionH.InsertMany(context.TODO(), newPoints)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Successfully inserted test data into museum DB")

	collectionR := client.Database("attractions-db").Collection("restaurants")
	_, err = collectionR.InsertMany(context.TODO(), newRestaurants)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Successfully inserted test data into restaurant DB")

	collectionM := client.Database("attractions-db").Collection("museums")
	_, err = collectionM.InsertMany(context.TODO(), newMuseums)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Successfully inserted test data into museum DB")

	return client, func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

}
