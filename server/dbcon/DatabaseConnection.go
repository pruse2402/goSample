package dbcon

import (
	"log"

	"gopkg.in/mgo.v2"
)

var dbSession = new(mgo.Session)

func ConnectMongoDB() {
	session, err := mgo.Dial("localhost:27017")

	if err != nil {
		log.Fatalf("Error in ping mongo: %s", err.Error())
	}

	session.SetMode(mgo.Monotonic, true)

	dbSession = session
}

func CopyDB() *mgo.Database {
	session := dbSession.Copy()
	return session.DB("gosample")
}

func CloseDB() {
	dbSession.Close()
}
