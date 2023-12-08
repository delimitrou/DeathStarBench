package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RoomType struct {
	BookableRate       float64 `bson:"bookableRate"`
	Code               string  `bson:"code"`
	RoomDescription    string  `bson:"roomDescription"`
	TotalRate          float64 `bson:"totalRate"`
	TotalRateInclusive float64 `bson:"totalRateInclusive"`
}

type RatePlan struct {
	HotelId  string    `bson:"hotelId"`
	Code     string    `bson:"code"`
	InDate   string    `bson:"inDate"`
	OutDate  string    `bson:"outDate"`
	RoomType *RoomType `bson:"roomType"`
}

func initializeDatabase(url string) (*mongo.Client, func()) {
	log.Info().Msg("Generating test data...")

	newRatePlans := []interface{}{
		RatePlan{
			"1",
			"RACK",
			"2015-04-09",
			"2015-04-10",
			&RoomType{
				109.00,
				"KNG",
				"King sized bed",
				109.00,
				123.17,
			},
		},
		RatePlan{
			"2",
			"RACK",
			"2015-04-09",
			"2015-04-10",
			&RoomType{
				139.00,
				"QN",
				"Queen sized bed",
				139.00,
				153.09,
			},
		},
		RatePlan{
			"3",
			"RACK",
			"2015-04-09",
			"2015-04-10",
			&RoomType{
				109.00,
				"KNG",
				"King sized bed",
				109.00,
				123.17,
			},
		},
	}

	for i := 7; i <= 80; i++ {
		if i%3 != 0 {
			continue
		}

		hotelID := strconv.Itoa(i)

		endDate := "2015-04-"
		if i%2 == 0 {
			endDate = fmt.Sprintf("%s17", endDate)
		} else {
			endDate = fmt.Sprintf("%s24", endDate)
		}

		rate := 109.00
		rateInc := 123.17
		if i%5 == 1 {
			rate = 120.00
			rateInc = 140.00
		} else if i%5 == 2 {
			rate = 124.00
			rateInc = 144.00
		} else if i%5 == 3 {
			rate = 132.00
			rateInc = 158.00
		} else if i%5 == 4 {
			rate = 232.00
			rateInc = 258.00
		}

		newRatePlans = append(
			newRatePlans,
			RatePlan{
				hotelID,
				"RACK",
				"2015-04-09",
				endDate,
				&RoomType{
					rate,
					"KNG",
					"King sized bed",
					rate,
					rateInc,
				},
			},
		)
	}

	uri := fmt.Sprintf("mongodb://%s", url)
	log.Info().Msgf("Attempting connection to %v", uri)

	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msg("Successfully connected to MongoDB")

	collection := client.Database("rate-db").Collection("inventory")
	_, err = collection.InsertMany(context.TODO(), newRatePlans)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Successfully inserted test data into rate DB")

	return client, func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
}
