package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

func initializeDatabase(url string) *mgo.Session {
	session, err := mgo.Dial(url)
	if err != nil {
		log.Panic().Msg(err.Error())
	}
	// defer session.Close()
	log.Info().Msg("New session successfull...")

	log.Info().Msg("Generating test data...")
	c := session.DB("user-db").C("user")
	for i := 0; i <= 500; i++ {
		suffix := strconv.Itoa(i)
		user_name := "Cornell_" + suffix
		password := ""
		for j := 0; j < 10; j++ {
			password += suffix
		}

		count, err := c.Find(&bson.M{"username": user_name}).Count()
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		if count == 0 {
			sum := sha256.Sum256([]byte(password))
			pass := fmt.Sprintf("%x", sum)
			err = c.Insert(&User{user_name, pass})
			if err != nil {
				log.Fatal().Msg(err.Error())
			}
		}

	}

	err = c.EnsureIndexKey("username")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	return session

	// count, err := c.Find(&bson.M{"username": "Cornell"}).Count()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if count == 0{
	// 	err = c.Insert(&User{"Cornell", "302eacf716390b1ebb39012b130302efec8a32ac4b8ad0a911112c53b60382b0"})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }

	// count, err = c.Find(&bson.M{"username": "ECE"}).Count()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if count == 0{
	// 	err = c.Insert(&User{"ECE", "a0a44ed8cfc32b7e61befeb99bbff7706808c3fe4dcdf4750a8addb3ffcd4008"})
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// }
}
