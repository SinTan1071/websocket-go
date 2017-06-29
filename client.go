package main

import (
	"too-white/log"
	// "bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"net/http"
	"time"
	// "too-white/conf"
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

func (serv *Server) read() {
	for {
		log.NewLog("client.go-39:首先阻塞", "")
		select {
		case message := <-serv.message:
			log.NewLog("client.go-42:读取到了消息", "")
			var m Request
			json.Unmarshal(message, &m)
			log.NewLog("client.go-45:传递过来的消息", string(message))
			var client_group []*Client
			var content Content
			log.NewLog("client.go-48:接收消息的用户tokens", m.Target)
			log.NewLog("client.go-49:查看一下所有的用户", serv.user)
			for _, user_token := range m.Target {
				if _, ok := serv.user[user_token]; ok {
					client_group = append(client_group, serv.user[user_token])
					content.Target = client_group
					log.NewLog("client.go-54:接收消息的用户（客户端）", client_group)
					log.NewLog("client.go-55:要发送的消息", m.Data)
					content.Data = m.Data
					serv.broadcast <- &content
				}
			}
			// default:
			// 	break
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
		log.NewLog("client.go-73:写阻塞开启", "")
		select {
		case message, ok := <-c.send:
			log.NewLog("client.go-76:写入消息", "")
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
			log.NewLog("client.go-96:写入心跳", "")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

func serveWsClient(serv *Server, w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	uid := r.Form.Get("uid")
	conn, err := upgrader.Upgrade(w, r, nil)
	log.NewLog("client.go-109:协议升级成功", "")
	if err != nil {
		log.NewLog("client.go-111:", err)
		return
	}
	client := &Client{serv: serv, conn: conn, send: make(chan []byte, 256), uid: uid}
	log.NewLog("client.go-115:开始客户端注册，第一步", "")
	client.serv.register <- client
	go client.writePump()
}

func serveWsServer(serv *Server, w http.ResponseWriter, r *http.Request) {
	go serv.read()
	err := r.ParseForm()
	// _target := r.Form.Get("target")
	// // _type := r.Form.Get("type")
	// _data := r.Form.Get("data")
	// fmt.Println("target", target)
	// fmt.Println("data", data)
	fmt.Println("数据", r.Body)
	target := []string{"0cc120124175285f80dc4b13e18730bb", "858803d74ab00f41c548f11ca4468df1"}
	request := Request{Target: target, Data: "lalal"}
	m, _ := json.Marshal(request)
	log.NewLog("client.go-46:②读取消息", "")
	serv.message <- m
	if err != nil {
		io.WriteString(w, "400")
	}
	io.WriteString(w, "200")
}
