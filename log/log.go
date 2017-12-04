package log

import (
	"TooWhite/conf"
	// "errors"
	"labix.org/v2/mgo"
	"strconv"
	"time"
)

type Log struct {
	Time    time.Time
	Comment string
	Content interface{}
}

func newLog() *mgo.Session {
	session, _ := mgo.Dial(conf.DB_DOMAIN + ":" + conf.DB_PORT)
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return session
}

func NewLog(comment string, data interface{}) {
	year := time.Now().Year()
	year_str := strconv.Itoa(year)
	month := time.Now().Month().String()
	log_name := year_str + "-" + month
	session := newLog()
	defer session.Close()
	log := &Log{
		Time:    time.Now(),
		Comment: comment,
		Content: data,
	}
	c := session.DB(conf.LOG_DATABASE).C(log_name)
	c.Insert(log)
}
