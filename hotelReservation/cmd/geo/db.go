package main

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type point struct {
	Pid  string  `bson:"hotelId"`
	Plat float64 `bson:"lat"`
	Plon float64 `bson:"lon"`
}

func initializeDatabase(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer session.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := session.DB("geo-db").C("geo")
	count, err := c.Find(&bson.M{"hotelId": "1"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&point{"1", 37.7867, -122.4112})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "2"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&point{"2", 37.7854, -122.4005})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "3"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&point{"3", 37.7854, -122.4071})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "4"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&point{"4", 37.7936, -122.3930})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "5"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&point{"5", 37.7831, -122.4181})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "6"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&point{"6", 37.7863, -122.4015})
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
		if count == 0 {
			err = c.Insert(&point{hotel_id, lat, lon})
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
