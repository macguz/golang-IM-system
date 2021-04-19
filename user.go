package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// 启动当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// Online 用户上线
func (u *User) Online() {
	// 用户上线，将用户加入onlineMap中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上限消息
	u.server.BroadCast(u, "已上线")
}

// Offline 用户下线
func (u *User) Offline() {
	// 用户下线，将用户从onlineMap中删除
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户上限消息
	u.server.BroadCast(u, "下线")
}

// SendMsg 发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))
}

// DoMessage 处理用户发送消息
func (u *User) DoMessage(msg string) {
	if msg == "who" {
		// 查询当前所有在线用户
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ": 在线..\n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		// 当前用户重命名
		newName := strings.Split(msg, "|")[1]
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("用户名已存在..\n")
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.Name = newName
			u.server.OnlineMap[u.Name] = u
			u.server.mapLock.Unlock()
			u.SendMsg("用户名已修改为: " + u.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {

		// 1. 取出用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			u.SendMsg("请输入用户名, 正确格式: \"to|<user>|<msg>\n\"")
			return
		}

		// 2. 提取用户
		remoteUser, ok := u.server.OnlineMap[remoteName]
		if !ok {
			u.SendMsg("用户名不存在\n")
		}

		// 3. 发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			u.SendMsg("无法发送空白消息\n")
			return
		}
		remoteUser.SendMsg(u.Name + " 对您说: " + content + "\n")
	} else {
		u.server.BroadCast(u, msg)
	}
}

// ListenMessage 监听当前User channel的方法，一旦有消息，就直接发送给客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
