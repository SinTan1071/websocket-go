package main

import (
	// "fmt"
	"TooWhite/log"
	"net/http"
)

type Request struct {
	MsgType int
	From    string
	Target  string
	Data    string
}
type Content struct {
	ContentType int
	From        *Client
	Target      []*Client
	Data        interface{}
}
type Response struct {
	Code int
	Msg  string
	Data interface{}
}

func main() {
	serv := newServer()
	go serv.run()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveWs(serv, w, r)
	})
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		log.NewLog("ListenAndServe-ERROR: ", err)
	}
}
