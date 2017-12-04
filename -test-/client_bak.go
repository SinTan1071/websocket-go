package main

import (
	"TooWhite/db"
	"TooWhite/helper"
	// "bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
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
				log.Printf("error: %v", err)
			}
			break
		}
		var m Msg
		err = json.Unmarshal(message, &m)
		fmt.Println("读取到的Msg结构体", m)
		if m.From != "" {
			if m.MsgType == 0 {
				// 加入用户进入在线状态
				c.uid = m.From
				user := &db.User{
					Name:  m.Data,
					Token: m.From,
				}
				user = db.UserJoin(user)
				broadcast_group := make([]*Client, 0)
				broadcast_group = append(broadcast_group, c)
				var content Content
				content.From = c
				content.Target = broadcast_group
				content.Data = "用户登录成功"
				c.serv.broadcast <- &content
				offline_msgs := db.GetUserOffLineMsg(m.From)
				for _, offline_msg := range offline_msgs {
					content.Data = offline_msg.Content
					c.serv.broadcast <- &content
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
				broadcast_group := make([]*Client, 0)
				broadcast_group = append(broadcast_group, c)
				var content Content
				content.From = c
				content.Target = broadcast_group
				content.Data = group
				c.serv.broadcast <- &content
			} else if m.MsgType == 2 {
				// 用户加入一个分组
				user := db.GetUserByToken(m.From)
				group := db.GetGroupByToken(m.Data)
				db.UserJoinGroup(m.From, m.Data)
				broadcast_group := make([]*Client, 0)
				var content Content
				content.From = c
				content.Data = "欢迎" + user.Name + "加入" + group.Name + "分组"
				for client, _ := range c.serv.clients {
					group = db.GetGroupByToken(m.Data)
					for _, user_token := range group.Users {
						if db.IsUserOnline(user_token) {
							// 用户在线
							if client.uid == user_token {
								broadcast_group = append(broadcast_group, client)
							}
							content.Target = broadcast_group
							c.serv.broadcast <- &content
						} else {
							// 用户离线
							offline_msg := &db.OffLineMsg{
								SendFrom: m.From,
								SendTo:   user_token,
								SendTime: time.Now(),
								Content:  "欢迎" + user.Name + "加入" + group.Name + "分组",
							}
							db.SaveUserOffLineMsg(offline_msg)
						}
					}
				}
			} else if m.MsgType == 3 {
				// 私聊
				from_user := db.GetUserByToken(m.From)
				to_user := db.GetUserByToken(m.Target)
				broadcast_group := make([]*Client, 0)
				broadcast_group = append(broadcast_group, c)
				var content Content
				content.From = c
				content.Data = from_user.Name + "对" + to_user.Name + "说：" + m.Data
				for client, _ := range c.serv.clients {
					if db.IsUserOnline(to_user.Token) {
						// 用户在线
						if client.uid == to_user.Token {
							broadcast_group = append(broadcast_group, client)
						}
						content.Target = broadcast_group
						c.serv.broadcast <- &content
					} else {
						// 用户离线
						offline_msg := &db.OffLineMsg{
							SendFrom: m.From,
							SendTo:   m.Target,
							SendTime: time.Now(),
							Content:  from_user.Name + "对" + to_user.Name + "说：" + m.Data,
						}
						db.SaveUserOffLineMsg(offline_msg)
					}

				}
			} else if m.MsgType == 4 {
				// 群聊
				from_user := db.GetUserByToken(m.From)
				to_group := db.GetGroupByToken(m.Target)
				broadcast_group := make([]*Client, 0)
				var content Content
				content.From = c
				content.Data = from_user.Name + "在群" + to_group.Name + "说：" + m.Data
				for client, _ := range c.serv.clients {
					for _, user_token := range to_group.Users {
						if db.IsUserOnline(user_token) {
							// 用户在线
							if client.uid == user_token {
								broadcast_group = append(broadcast_group, client)
							}
							content.Target = broadcast_group
							c.serv.broadcast <- &content
						} else {
							// 用户离线
							offline_msg := &db.OffLineMsg{
								SendFrom: m.From,
								SendTo:   user_token,
								SendTime: time.Now(),
								Content:  from_user.Name + "在群" + to_group.Name + "说：" + m.Data,
							}
							db.SaveUserOffLineMsg(offline_msg)
						}
					}
				}
			} else if m.MsgType == 5 {
				// 删除分组
				broadcast_group := make([]*Client, 0)
				broadcast_group = append(broadcast_group, c)
				var content Content
				content.From = c
				content.Target = broadcast_group
				if db.DelGroup(m.From, m.Data) {
					content.Data = "删除成功"
				} else {
					content.Data = "删除失败"
				}
				c.serv.broadcast <- &content
			} else if m.MsgType == 6 {
				// 用户离开分组
				user := db.GetUserByToken(m.From)
				group := db.GetGroupByToken(m.Data)
				db.UserOffGroup(m.From, m.Data)
				broadcast_group := make([]*Client, 0)
				var content Content
				content.From = c
				content.Data = user.Name + "离开了" + group.Name + "分组"
				for client, _ := range c.serv.clients {
					for _, user_token := range group.Users {
						if db.IsUserOnline(user_token) {
							// 用户在线
							if client.uid == user_token {
								broadcast_group = append(broadcast_group, client)
							}
							content.Target = broadcast_group
							c.serv.broadcast <- &content
						} else {
							// 用户离线
							offline_msg := &db.OffLineMsg{
								SendFrom: m.From,
								SendTo:   user_token,
								SendTime: time.Now(),
								Content:  user.Name + "离开了" + group.Name + "分组",
							}
							db.SaveUserOffLineMsg(offline_msg)
						}
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
		log.Println(err)
		return
	}
	client := &Client{serv: serv, conn: conn, send: make(chan []byte, 256)}
	client.serv.register <- client
	go client.writePump()
	client.readPump()
}