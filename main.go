package main

import (
	"net/http"
	"too-white/conf"
	"too-white/log"
)

type Request struct {
	Type   int
	From   string
	Target []string
	Data   string
}
type Content struct {
	From   *Client
	Target []*Client
	Data   interface{}
}

func main() {
	serv := newServer()
	go serv.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(serv, w, r)
	})
	err := http.ListenAndServe(conf.PORT, nil)
	if err != nil {
		log.NewLog("ListenAndServe-ERROR: ", err)
	}
}
