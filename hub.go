package main

import (
	"TooWhite/db"
	"TooWhite/log"
	"encoding/json"
)

type Server struct {
	clients map[*Client]bool

	broadcast chan *Content

	register chan *Client

	unregister chan *Client
}

func newServer() *Server {
	return &Server{
		broadcast:  make(chan *Content),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (serv *Server) run() {
	for {
		select {
		case client := <-serv.register:
			serv.clients[client] = true
			log.NewLog("hub.go-33:把客户端放在线程池里", serv.clients)
		case client := <-serv.unregister:
			if _, ok := serv.clients[client]; ok {
				delete(serv.clients, client)
				close(client.send)
				db.UserOffLine(client.uid)
			}
		case content := <-serv.broadcast:
			// 根据content传过来的需要广播的target来遍历广播
			message, _ := json.Marshal(content.Data)
			if content.Target != nil {
				for _, client := range content.Target {
					if _, ok := serv.clients[client]; ok {
						select {
						case client.send <- message:
						default:
							close(client.send)
							delete(serv.clients, client)
						}
					}
				}
			}
		}
	}
}
