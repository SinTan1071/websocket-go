package main

import (
	"too-white/log"
	// "bytes"
	"encoding/json"
	// "fmt"
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
		log.NewLog("client.go-46:", "进入readPump的for循环（开始阻塞）")
		_, message, err := c.conn.ReadMessage()
		log.NewLog("client.go-48:", "进入readPump的for循环读取数据")
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.NewLog("client.go-51:", err)
			}
			break
		}
		var m Request
		json.Unmarshal(message, &m)
		log.NewLog("client.go-57:传递过来的消息", string(message))
		log.NewLog("client.go-58:读取到的Msg结构体", m)
		if m.Type == conf.REQUEST_TYPE_CLIENT {
			c.uid = m.From
			var client_group []*Client
			var content Content
			content.From = c
			client_group = append(client_group, c)
			content.Target = client_group
			log.NewLog("client.go-66:接收消息的用户（客户端）", client_group)
			log.NewLog("client.go-67:要发送的消息", m.Data)
			content.Data = "连接成功"
			c.serv.broadcast <- &content
		} else if m.Type == conf.REQUEST_TYPE_SERVER {
			if m.From == conf.SERVER_TOKEN {
				var client_group []*Client
				var content Content
				content.From = c
				log.NewLog("client.go-75:接收消息的用户tokens", m.Target)
				for client, _ := range c.serv.clients {
					for _, user_token := range m.Target {
						if client.uid == user_token {
							client_group = append(client_group, client)
							content.Target = client_group
							log.NewLog("client.go-81:接收消息的用户（客户端）", client_group)
							log.NewLog("client.go-82:要发送的消息", m.Data)
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
		log.NewLog("client.go-102:写阻塞开启", "")
		select {
		case message, ok := <-c.send:
			log.NewLog("client.go-105:写入消息", "")
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
			log.NewLog("client.go-125:写入心跳", "")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWs(serv *Server, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	log.NewLog("client.go-136:协议升级成功", err)
	if err != nil {
		log.NewLog("client.go-138:", err)
		return
	}
	client := &Client{serv: serv, conn: conn, send: make(chan []byte, 256)}
	log.NewLog("client.go-142:开始客户端注册，第一步", "")
	client.serv.register <- client
	go client.writePump()
	client.readPump()
}
