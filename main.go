package main

import (
	"net/http"
	"too-white/conf"
	"too-white/log"
)

type Request struct {
	Target []string
	Data   string
}
type Content struct {
	// From   *Client
	Target []*Client
	Data   interface{}
}

var check = make(chan bool, 1024)

func main() {
	serv := newServer()
	go serv.run()
	http.HandleFunc("/c", func(w http.ResponseWriter, r *http.Request) {
		serveWsClient(serv, w, r)
	})
	http.HandleFunc("/s", func(w http.ResponseWriter, r *http.Request) {
		serveWsServer(serv, w, r)
	})
	err := http.ListenAndServe(conf.PORT, nil)
	if err != nil {
		log.NewLog("ListenAndServe-ERROR: ", err)
	}
}
