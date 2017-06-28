package main

import (
	"encoding/json"
	"too-white/log"
)

type Server struct {
	clients map[*Client]bool
	user    map[string]*Client
	// group           map[string]map[*Client]bool
	// room            map[string]map[*Client]bool
	broadcast chan *Content
	// broadcast2user  chan *Content
	// broadcast2group chan *Content
	// broadcast2room  chan *Content
	register   chan *Client
	unregister chan *Client
	message    chan []byte
}

func newServer() *Server {
	return &Server{
		broadcast: make(chan *Content),
		// broadcast2group: make(chan *Content),
		// broadcast2room:  make(chan *Content),
		// broadcast2user:  make(chan *Content),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		user:       make(map[string]*Client),
		message:    make(chan []byte),
		// group:           make(map[string]map[*Client]bool),
		// room:            make(map[string]map[*Client]bool),
	}
}

func (serv *Server) run() {
	for {
		log.NewLog("hub.go-26:第一个goroutine for 循环阻塞", "")
		select {
		case client := <-serv.register:
			serv.clients[client] = true
			serv.user[client.uid] = client
			log.NewLog("hub.go-30:把客户端放在线程池里，第二步", serv.clients)
		case client := <-serv.unregister:
			log.NewLog("hub.go-32:用户离线", serv.clients)
			if _, ok := serv.clients[client]; ok {
				delete(serv.clients, client)
				close(client.send)
				// db.UserOffLine(client.uid)
			}
		case content := <-serv.broadcast:
			// 根据content传过来的需要广播的target来遍历广播
			log.NewLog("hub.go-40:开始广播了！！！", "")
			message, _ := json.Marshal(content.Data)
			log.NewLog("发送的消息是是是是是是", string(message))
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
