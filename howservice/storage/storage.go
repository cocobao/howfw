package storage

import (
	"os"

	"github.com/cocobao/log"
	"gopkg.in/mgo.v2"
)

var (
	mongo *mgo.Session
)

func SetupMongoDB(url string) {
	var err error
	mongo, err = mgo.Dial(url)
	if err != nil {
		log.Error("dial mongo db fail", err)
		os.Exit(0)
	}
	log.Info("setup mongo db success")
}

func InsertEventData(data interface{}) {
	session := mongo.Clone()
	defer session.Close()
	c := session.DB("log_center").C("device_event")

	c.Insert(data)
}
