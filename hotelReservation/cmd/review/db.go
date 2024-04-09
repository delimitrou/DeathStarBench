package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Review struct {
	ReviewId    string  `bson:"reviewId"`
	HotelId     string  `bson:"hotelId"`
	Name        string  `bson:"name"`
	Rating      float32 `bson:"rating"`
	Description string  `bson:"description"`
	Image       *Image  `bson:"images"`
}

type Image struct {
	Url     string `bson:"url"`
	Default bool   `bson:"default"`
}

func initializeDatabase(url string) (*mongo.Client, func()) {

	newReviews := []interface{}{
		&Review{
			"1",
			"1",
			"Person 1",
			3.4,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}},
		&Review{
			"2",
			"1",
			"Person 2",
			4.4,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}},
		&Review{
			"3",
			"1",
			"Person 3",
			4.2,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}},
		&Review{
			"4",
			"1",
			"Person 4",
			3.9,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}},
		&Review{
			"5",
			"2",
			"Person 5",
			4.2,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}},
		&Review{
			"6",
			"2",
			"Person 6",
			3.7,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}},
	}

	uri := fmt.Sprintf("mongodb://%s", url)
	log.Info().Msgf("Attempting connection to %v", uri)

	opts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	log.Info().Msg("Successfully connected to MongoDB")

	collection := client.Database("review-db").Collection("reviews")
	_, err = collection.InsertMany(context.TODO(), newReviews)
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
