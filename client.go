package main

import (
	"too-white/log"
	// "bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
	"too-white/conf"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Client struct {
	serv *Server
	conn *websocket.Conn
	send chan []byte
	uid  string //存放用户的唯一标示
}

func (c *Client) readPump() {
	defer func() {
		c.serv.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	log.NewLog("client.go-44:", "进入readPump")
	for {
		log.NewLog("client.go-46:", "进入readPump的for循环")
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.NewLog("client.go-48:", err)
			}
			break
		}
		var m Request
		json.Unmarshal(message, &m)
		fmt.Println("传递过来的消息", string(message))
		fmt.Println("传递过来的消息", m)
		log.NewLog("client.go-54:读取到的Msg结构体", m)
		if m.Type == conf.REQUEST_TYPE_CLIENT {
			c.uid = m.From
			var client_group []*Client
			var content Content
			content.From = c
			client_group = append(client_group, c)
			content.Target = client_group
			log.NewLog("client.go-62:接收消息的用户（客户端）", client_group)
			log.NewLog("client.go-63:要发送的消息", m.Data)
			content.Data = "连接成功"
			c.serv.broadcast <- &content
		} else if m.Type == conf.REQUEST_TYPE_SERVER {
			if m.From == conf.SERVER_TOKEN {
				var client_group []*Client
				var content Content
				content.From = c
				log.NewLog("client.go-71:接收消息的用户tokens", m.Target)
				for client, _ := range c.serv.clients {
					for _, user_token := range m.Target {
						if client.uid == user_token {
							client_group = append(client_group, client)
							content.Target = client_group
							log.NewLog("client.go-77:接收消息的用户（客户端）", client_group)
							log.NewLog("client.go-78:要发送的消息", m.Data)
							content.Data = m.Data
							c.serv.broadcast <- &content
						}
					}
				}
			} else {
				return
			}
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			n := len(c.send)
			for i := 0; i < n; i++ {
				// w.Write(newline)
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(serv *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.NewLog("client.go-130:", err)
		return
	}
	client := &Client{serv: serv, conn: conn, send: make(chan []byte, 256)}
	client.serv.register <- client
	go client.writePump()
	client.readPump()
}
