package main

import (
	"TooWhite/db"
	"TooWhite/helper"
	"TooWhite/log"
	// "bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

const (
	writeWait = 10 * time.Second

	pongWait = 60 * time.Second

	pingPeriod = (pongWait * 9) / 10

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

	uid string //存放用户的唯一标示

}

func (c *Client) readPump() {
	defer func() {
		c.serv.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.NewLog("client.go-56:", err)
			}
			break
		}
		var m Request
		var res Response
		var user_token_group []string
		json.Unmarshal(message, &m)
		log.NewLog("client.go-64:读取到的Msg结构体", m)
		if m.From != "" {
			if m.MsgType == 0 {
				// 加入用户进入在线状态
				c.uid = m.From
				user := &db.User{
					Name:  m.Data,
					Token: m.From,
				}
				user = db.UserJoin(user)
				user_token_group = append(user_token_group, m.From)
				offline_msgs := db.GetUserOffLineMsg(m.From)
				res.Msg = "用户登录成功"
				res.Data = offline_msgs
				for range offline_msgs {
					db.DelUserOffLineMsg(m.From)
				}
			} else if m.MsgType == 1 {
				// 增加一个分组
				gtoken := helper.MakeGroupToken(m.From)
				user := db.GetUserByToken(m.From)
				group := &db.Group{
					Name:    m.Data,
					Token:   gtoken,
					Creater: user.Token,
					Users:   []string{user.Token},
				}
				db.NewGroup(group)
				user_token_group = append(user_token_group, m.From)
				res.Msg = "分组创建成功"
				res.Data = group
			} else if m.MsgType == 2 {
				// 用户加入一个分组
				user := db.GetUserByToken(m.From)
				group := db.GetGroupByToken(m.Data)
				db.UserJoinGroup(m.From, m.Data)
				res.Msg = "欢迎" + user.Name + "加入" + group.Name + "分组"
				group = db.GetGroupByToken(m.Data)
				user_token_group = group.Users
			} else if m.MsgType == 3 {
				// 私聊
				from_user := db.GetUserByToken(m.From)
				to_user := db.GetUserByToken(m.Target)
				res.Msg = from_user.Name + "对" + to_user.Name + "说：" + m.Data
				//
				// 使用redis获取名称
				// from_user_name := db.GetUserNameByToken(m.From)
				// to_user_name := db.GetUserNameByToken(m.Target)
				// res.Msg = from_user_name + "对" + to_user_name + "说：" + m.Data
				//
				user_token_group = []string{m.From, m.Target}
			} else if m.MsgType == 4 {
				// 群聊
				from_user := db.GetUserByToken(m.From)
				to_group := db.GetGroupByToken(m.Target)
				res.Msg = from_user.Name + "在群" + to_group.Name + "说：" + m.Data
				//
				// 使用redis获取名称
				// from_user_name := db.GetUserNameByToken(m.From)
				// to_group_name := db.GetGroupNameByToken(m.Target)
				// to_group := db.GetGroupUsersByToken(m.Target)
				// res.Msg = from_user_name + "在群" + to_group_name + "说：" + m.Data
				//
				user_token_group = to_group.Users
			} else if m.MsgType == 5 {
				// 删除分组
				user_token_group = append(user_token_group, m.From)
				if db.DelGroup(m.From, m.Data) {
					res.Msg = "删除成功"
				} else {
					res.Msg = "删除失败"
				}
			} else if m.MsgType == 6 {
				// 用户离开分组
				user := db.GetUserByToken(m.From)
				group := db.GetGroupByToken(m.Data)
				db.UserOffGroup(m.From, m.Data)
				res.Msg = user.Name + "离开了" + group.Name + "分组"
				user_token_group = group.Users
			}
			var client_group []*Client
			var content Content
			content.From = c
			log.NewLog("client.go-133:接收消息的用户tokens", user_token_group)
			for client, _ := range c.serv.clients {
				for _, user_token := range user_token_group {
					if db.IsUserOnline(user_token) {
						// 用户在线
						if client.uid == user_token {
							client_group = append(client_group, client)
							content.Target = client_group
							log.NewLog("client.go-141:接收消息的用户（客户端）", client_group)
							log.NewLog("client.go-142:要发送的消息", res)
							content.Data = res
							c.serv.broadcast <- &content
						}
					} else {
						// 用户离线
						offline_msg := &db.OffLineMsg{
							SendFrom: m.From,
							SendTo:   user_token,
							SendTime: time.Now(),
							Content:  res,
						}
						db.SaveUserOffLineMsg(offline_msg)
					}
				}
			}
		} else {
			return
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
		log.NewLog("client.go-207:", err)
		return
	}
	client := &Client{serv: serv, conn: conn, send: make(chan []byte, 256)}
	client.serv.register <- client
	go client.writePump()
	client.readPump()
}
