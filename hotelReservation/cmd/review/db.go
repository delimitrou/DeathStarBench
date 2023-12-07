package main

import (
	"github.com/rs/zerolog/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"fmt"
)

type Review struct {
	ReviewId	string  `bson:"reviewId"`
	HotelId		 string  `bson:"hotelId"`
	Name         string  `bson:"name"`
	Rating       float32 `bson:"rating"`
	Description  string  `bson:"description"`
	Image        *Image  `bson:"images"`
}

type Image struct {
	Url		 string  `bson:"url"`
	Default  bool    `bson:"default"`
}

func initializeDatabase(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	fmt.Println("Initialize Database ", url)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer session.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := session.DB("review-db").C("reviews")
	count, err := c.Find(&bson.M{"reviewId": "1"}).Count()
	fmt.Println("Count = ", count)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		fmt.Println("Insert 1")
		err = c.Insert(&Review{
			"1",
			"1",
			"Person 1",
			3.4,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"reviewId": "2"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		fmt.Println("Insert 2")
		err = c.Insert(&Review{
			"2",
			"1",
			"Person 2",
			4.4,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"reviewId": "3"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		fmt.Println("Insert 3")
		err = c.Insert(&Review{
			"3",
			"1",
			"Person 3",
			4.2,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"reviewId": "4"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		fmt.Println("Insert 4")
		err = c.Insert(&Review{
			"4",
			"1",
			"Person 4",
			3.9,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"reviewId": "5"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		fmt.Println("Insert 5")
		err = c.Insert(&Review{
			"5",
			"2",
			"Person 5",
			4.2,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	count, err = c.Find(&bson.M{"reviewId": "6"}).Count()
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	if count == 0 {
		fmt.Println("Insert 6")
		err = c.Insert(&Review{
			"6",
			"2",
			"Person 6",
			3.7,
			"A 6-minute walk from Union Square and 4 minutes from a Muni Metro station, this luxury hotel designed by Philippe Starck features an artsy furniture collection in the lobby, including work by Salvador Dali.",
			&Image{
				"some url",
				false}})
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
	}

	err = c.EnsureIndexKey("reviewId")
	err = c.EnsureIndexKey("hotelId")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	return session
}
