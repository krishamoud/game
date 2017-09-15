// Package db manages the connections to the database
package db

import (
	"math/rand"
	"os"
	"time"

	"gopkg.in/mgo.v2"
)

// MongoConn is the connection to the MongoDB that we keep open
var MongoConn = getConn()

// DB is the database that we connect to
var DB = getDB(MongoConn)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func getConn() *mgo.Session {
	var err error
	mongoConn, err := mgo.Dial(os.Getenv("MONGO_URL"))
	if err != nil {
		panic(err)
	}
	return mongoConn
}

func getDB(mongoConn *mgo.Session) *mgo.Database {
	// Get the db name from the mongo env variable
	dialInfo, _ := mgo.ParseURL(os.Getenv("MONGO_URL"))
	dB := mongoConn.DB(dialInfo.Database)
	return dB
}

// RandomID generates a new random id of n length
func RandomID(n int) string {
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
