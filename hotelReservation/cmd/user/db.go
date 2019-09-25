package main

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"strconv"
	"log"
	"fmt"
)

type User struct {
	Username string `bson:"username"`
	Password string `bson:"password"`
}

func initializeDatabase(url string) *mgo.Session {
	fmt.Printf("user db ip addr = %s\n", url)
	session, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	// defer session.Close()

	c := session.DB("user-db").C("user")

	for i := 0; i <= 500; i++ {
		suffix := strconv.Itoa(i)
		user_name := "Cornell_" + suffix
		password  := ""
		for j := 0; j < 10; j++ {
			password += suffix
		}

		count, err := c.Find(&bson.M{"username": user_name}).Count()
		if err != nil {
			log.Fatal(err)
		}
		if count == 0{
			err = c.Insert(&User{user_name, password})
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	err = c.EnsureIndexKey("username")
	if err != nil {
		log.Fatal(err)
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