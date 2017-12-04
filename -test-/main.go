package main

import (
	"fmt"
	"golang.org/x/net/websocket"
	// "io"
	"log"
	"net/http"
)

func ChatWith(ws *websocket.Conn) {
	var err error

	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println(" Can't receive ")
			break
		}

		fmt.Println(" Received back from client: " + reply)

		// msg := "Received from " + ws.Request().Host + "  " + reply
		msg := " welcome to websocket do by pp "
		fmt.Println(" Sending to client: " + msg)

		if err = websocket.Message.Send(ws, msg); err != nil {
			fmt.Println(" Can't send ")
			break
		}
	}
}

func main() {
	//
	http.Handle("/", websocket.Handler(ChatWith))

	fmt.Println(" listen on port 9999 ")

	if err := http.ListenAndServe(":9999 ", nil); err != nil {
		log.Fatal(" ListenAndServe: ", err)
	}
}
