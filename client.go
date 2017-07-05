package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"too-white/conf"
	"too-white/log"
	"too-white/util"
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
	app  string //存放用户的唯一标示
}

func (serv *Server) read() {
	for {
		log.NewLog("client.go-41:读取read（）首先阻塞", "")
		select {
		case req := <-serv.request:
			log.NewLog("client.go-44:传递过来的消息", req)
			var client_group []*Client
			var content Content
			log.NewLog("client.go-47:接收消息的用户tokens", req.Target)
			log.NewLog("client.go-48:查看一下所有的用户", serv.user)
			for _, uid := range req.Target {
				for _, app := range conf.CLIENT_APP {
					if _, ok := serv.user[uid+":"+app]; ok {
						client_group = append(client_group, serv.user[uid+":"+app])
					}
				}
			}
			content.Target = client_group
			log.NewLog("client.go-57:接收消息的用户（客户端）", client_group)
			log.NewLog("client.go-58:要发送的消息", req.Data)
			content.Data = req.Data
			serv.broadcast <- &content
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
		log.NewLog("client.go-74:写阻塞开启", "")
		select {
		case message, ok := <-c.send:
			log.NewLog("client.go-77:写入消息", "")
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
				w.Write(<-c.send)
			}
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			log.NewLog("client.go-96:写入心跳", "")
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.NewLog("写入心跳不成功，无效的链接---", c.uid+":"+c.app)
				return
			}
		}
	}
}

func serveWsClient(serv *Server, w http.ResponseWriter, r *http.Request) {
	log.NewLog("client.go-106:所有的客户端", serv.user)
	r.ParseForm()
	uid := r.Form.Get("uid")
	token := r.Form.Get("token")
	app := r.Form.Get("app")
	go Auth(uid, token)
	ok := <-check
	if !ok || uid == "" || token == "" || app == "" {
		io.WriteString(w, "400")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	log.NewLog("client.go-118:协议升级成功", "")
	if err != nil {
		log.NewLog("client.go-120:", err)
		return
	}
	// 踢出同一个账号的另外一个长连接
	if _, ok := serv.clients[serv.user[uid+":"+app]]; ok {
		log.NewLog("client.go-125:踢出同一个账号的另外一个长连接", serv.user[uid+":"+app])
		delete(serv.clients, serv.user[uid+":"+app])
		close((serv.user[uid+":"+app]).send)
	}
	if _, ok := serv.user[uid+":"+app]; ok {
		delete(serv.user, uid+":"+app)
	}
	client := &Client{serv: serv, conn: conn, send: make(chan []byte, 256), uid: uid, app: app}
	log.NewLog("client.go-133:开始客户端注册，第一步", "")
	client.serv.register <- client
	go client.writePump()
}

func serveWsServer(serv *Server, w http.ResponseWriter, r *http.Request) {
	ip := util.GetIp()
	log.NewLog("client.go-服务端的IP-140:", ip)
	if ip != conf.SERVER_IP {
		io.WriteString(w, "400")
		return
	}
	go serv.read()
	err1 := r.ParseForm()
	log.NewLog("client.go-服务端消息验证-147:", err1)
	_msg := r.Form.Get("Message")
	log.NewLog("client.go-149:读取服务器消息", _msg)
	_count := r.Form.Get("Count")
	if _msg != "" && err1 == nil {
		var req Request
		err2 := json.Unmarshal([]byte(_msg), &req)
		log.NewLog("client.go-154:读取服务器消息解析后的", req)
		if err2 != nil || &req == nil {
			log.NewLog("client.go-服务端消息验证156:", err2)
			io.WriteString(w, "400")
			return
		}
		serv.request <- &req
		io.WriteString(w, "200")
		return
	}
	if _count == "online_users" {
		log.NewLog("!当前在线人!", serv.user)
		io.WriteString(w, "当前在线人数"+fmt.Sprint(len(serv.user)))
		return
	}
	io.WriteString(w, "400")
	return
}

func Auth(uid, token string) {
	log.NewLog("client.go-Auth方法验证UID和Token:", uid+"---"+token)
	url := conf.CLIENT_AUTH + "?uid=" + uid + "&token=" + token
	resp, err := http.Get(url)
	if err != nil {
		log.NewLog("client.go-Auth方法请求服务器报错", err)
		check <- false
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	log.NewLog("服务器鉴权返回", string(body))
	if string(body) != "200" {
		check <- false
		return
	}
	check <- true
	return
}
