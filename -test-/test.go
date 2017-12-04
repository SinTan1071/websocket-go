package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		fmt.Println("===this is messageType===", messageType)
		fmt.Println("===this is p===", p)
		if messageType == 1 {
			fmt.Println("===sososos===", string(p))
		}
		// if err = conn.WriteMessage(messageType, p); err != nil {
		// 	return err
		// }
	}
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r)
	})
	if err := http.ListenAndServe(":9999", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
