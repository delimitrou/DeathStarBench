package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Hotel struct {
	Id          string   `bson:"id"`
	Name        string   `bson:"name"`
	PhoneNumber string   `bson:"phoneNumber"`
	Description string   `bson:"description"`
	Address     *Address `bson:"address"`
}

type Address struct {
	StreetNumber string  `bson:"streetNumber"`
	StreetName   string  `bson:"streetName"`
	City         string  `bson:"city"`
	State        string  `bson:"state"`
	Country      string  `bson:"country"`
	PostalCode   string  `bson:"postalCode"`
	Lat          float32 `bson:"lat"`
	Lon          float32 `bson:"lon"`
}

func initializeDatabase(url string) (*mongo.Client, func()) {
	log.Info().Msg("Generating test data...")

	newProfiles := []interface{}{
		Hotel{
			"1",
			"Clift Hotel",
			"(415) 775-4700",
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Address{
				"495",
				"Geary St",
				"San Francisco",
				"CA",
				"United States",
				"94102",
				37.7867,
				-122.4112,
			},
		},
		Hotel{
			"2",
			"W San Francisco",
			"(415) 777-5300",
			"Less than a block from the Yerba Buena Center for the Arts, this trendy hotel is a 12-minute walk from Union Square.",
			&Address{
				"181",
				"3rd St",
				"San Francisco",
				"CA",
				"United States",
				"94103",
				37.7854,
				-122.4005,
			},
		},
		Hotel{
			"3",
			"Hotel Zetta",
			"(415) 543-8555",
			"A 3-minute walk from the Powell Street cable-car turnaround and BART rail station, this hip hotel 9 minutes from Union Square combines high-tech lodging with artsy touches.",
			&Address{
				"55",
				"5th St",
				"San Francisco",
				"CA",
				"United States",
				"94103",
				37.7834,
				-122.4071,
			},
		},
		Hotel{
			"4",
			"Hotel Vitale",
			"(415) 278-3700",
			"This waterfront hotel with Bay Bridge views is 3 blocks from the Financial District and a 4-minute walk from the Ferry Building.",
			&Address{
				"8",
				"Mission St",
				"San Francisco",
				"CA",
				"United States",
				"94105",
				37.7936,
				-122.3930,
			},
		},
		Hotel{
			"5",
			"Phoenix Hotel",
			"(415) 776-1380",
			"Located in the Tenderloin neighborhood, a 10-minute walk from a BART rail station, this retro motor lodge has hosted many rock musicians and other celebrities since the 1950s. It’s a 4-minute walk from the historic Great American Music Hall nightclub.",
			&Address{
				"601",
				"Eddy St",
				"San Francisco",
				"CA",
				"United States",
				"94109",
				37.7831,
				-122.4181,
			},
		},
		Hotel{
			"6",
			"St. Regis San Francisco",
			"(415) 284-4000",
			"St. Regis Museum Tower is a 42-story, 484 ft skyscraper in the South of Market district of San Francisco, California, adjacent to Yerba Buena Gardens, Moscone Center, PacBell Building and the San Francisco Museum of Modern Art.",
			&Address{
				"125",
				"3rd St",
				"San Francisco",
				"CA",
				"United States",
				"94109",
				37.7863,
				-122.4015,
			},
		},
	}

	for i := 7; i <= 80; i++ {
		hotelID := strconv.Itoa(i)
		phoneNumber := fmt.Sprintf("(415) 284-40%s", hotelID)

		lat := 37.7835 + float32(i)/500.0*3
		lon := -122.41 + float32(i)/500.0*4

		newProfiles = append(
			newProfiles,
			Hotel{
				hotelID,
				"St. Regis San Francisco",
				phoneNumber,
				"St. Regis Museum Tower is a 42-story, 484 ft skyscraper in the South of Market district of San Francisco, California, adjacent to Yerba Buena Gardens, Moscone Center, PacBell Building and the San Francisco Museum of Modern Art.",
				&Address{
					"125",
					"3rd St",
					"San Francisco",
					"CA",
					"United States",
					"94109",
					lat,
					lon,
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

	collection := client.Database("profile-db").Collection("hotels")
	_, err = collection.InsertMany(context.TODO(), newProfiles)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	log.Info().Msg("Successfully inserted test data into profile DB")

	return client, func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
}
