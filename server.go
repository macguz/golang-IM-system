package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip	 string
	Port int
	// 在线用户列表
	OnlineMap map[string]*User
	mapLock	sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

// LisenMessager 监听Message广播消息channel的goroutine，一旦有消息就发送给所有在线User
func (s *Server) LisenMessager()  {
	for {
		msg := <-s.Message

		// 将msg发送给全部的在线User
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// BroadCast 广播消息
func (s *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

// Handler 处理当前连接
func (s *Server) Handler(conn net.Conn) {
	user := NewUser(conn, s)

	// 用户上线
	user.Online()

	// 监听当前用户是否活跃
	isALive := make(chan bool)

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 用户下线
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read error:", err)
			}

			// 提取用户消息
			msg := string(buf)[:n-1]

			// 将得到消息广播
			user.DoMessage(msg)

			isALive <- true
		}
	}()

	for {
		select {

		case <-isALive:

		case <-time.After(time.Minute * 5):
			user.SendMsg("你被踢了\n")
			// 销毁user的资源
			conn.Close()
			close(user.C)
			return
		}
	}
}


func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net listen error:", err)
		return
	}
	defer listener.Close()

	// 启动监听Message的goroutine
	go s.LisenMessager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error:", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}
}