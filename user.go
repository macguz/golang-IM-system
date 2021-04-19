package main

import "net"

type User struct {
	Name string
	Addr string
	C	 chan string
	conn net.Conn
	server 	 *Server
}

// NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
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
func (u *User) DoMessage(msg string)  {
	if msg == "who" {
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ": 在线..\n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
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
