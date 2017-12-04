package storage

import (
	"os"

	"github.com/cocobao/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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

func GetNewlestDevLog(devId string) (map[string]interface{}, error) {
	session := mongo.Clone()
	defer session.Close()
	c := session.DB("log_center").C("device_event")

	var result map[string]interface{}
	err := c.Find(bson.M{"device_id": devId}).Sort("-insert_time").Limit(1).One(&result)
	return result, err
}
