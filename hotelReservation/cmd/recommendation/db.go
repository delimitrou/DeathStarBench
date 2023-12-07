package main

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Hotel struct {
	HId    string  `bson:"hotelId"`
	HLat   float64 `bson:"lat"`
	HLon   float64 `bson:"lon"`
	HRate  float64 `bson:"rate"`
	HPrice float64 `bson:"price"`
}

func initializeDatabase(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer session.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := session.DB("recommendation-db").C("recommendation")
	count, err := c.Find(&bson.M{"hotelId": "1"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&Hotel{"1", 37.7867, -122.4112, 109.00, 150.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "2"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&Hotel{"2", 37.7854, -122.4005, 139.00, 120.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "3"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&Hotel{"3", 37.7834, -122.4071, 109.00, 190.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "4"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&Hotel{"4", 37.7936, -122.3930, 129.00, 160.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "5"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&Hotel{"5", 37.7831, -122.4181, 119.00, 140.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "6"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&Hotel{"6", 37.7863, -122.4015, 149.00, 200.00})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		hotel_id := strconv.Itoa(i)
		count, err = c.Find(&bson.M{"hotelId": hotel_id}).Count()
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		lat := 37.7835 + float64(i)/500.0*3
		lon := -122.41 + float64(i)/500.0*4

		count, err = c.Find(&bson.M{"hotelId": hotel_id}).Count()
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
			err = c.Insert(&Hotel{hotel_id, lat, lon, rate, rate_inc})
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
		}

	}

	err = c.EnsureIndexKey("hotelId")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	return session
}
