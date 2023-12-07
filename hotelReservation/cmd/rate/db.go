package main

import (
	"strconv"

	"github.com/rs/zerolog/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

func initializeDatabase(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer session.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := session.DB("rate-db").C("inventory")
	count, err := c.Find(&bson.M{"hotelId": "1"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&RatePlan{
			"1",
			"RACK",
			"2015-04-09",
			"2015-04-10",
			&RoomType{
				109.00,
				"KNG",
				"King sized bed",
				109.00,
				123.17}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "2"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&RatePlan{
			"2",
			"RACK",
			"2015-04-09",
			"2015-04-10",
			&RoomType{
				139.00,
				"QN",
				"Queen sized bed",
				139.00,
				153.09}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"hotelId": "3"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		err = c.Insert(&RatePlan{
			"3",
			"RACK",
			"2015-04-09",
			"2015-04-10",
			&RoomType{
				109.00,
				"KNG",
				"King sized bed",
				109.00,
				123.17}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	// add up to 80 hotels
	for i := 7; i <= 80; i++ {
		if i%3 == 0 {
			hotel_id := strconv.Itoa(i)
			count, err = c.Find(&bson.M{"hotelId": hotel_id}).Count()
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
			end_date := "2015-04-"
			rate := 109.00
			rate_inc := 123.17
			if i%2 == 0 {
				end_date = end_date + "17"
			} else {
				end_date = end_date + "24"
			}

			if i%5 == 1 {
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

			if count == 0 {
				err = c.Insert(&RatePlan{
					hotel_id,
					"RACK",
					"2015-04-09",
					end_date,
					&RoomType{
						rate,
						"KNG",
						"King sized bed",
						rate,
						rate_inc}})
				if err != nil {
					log.Fatal().Msg(err.Error())
				}
			}
		}
	}

	err = c.EnsureIndexKey("hotelId")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	return session
}
