package main

import (
	"fmt"
	"io"
	"net"
	"sync"
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
	user := NewUser(conn)

	// 用户上线，将用户加入onlineMap中
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// 广播当前用户上限消息
	s.BroadCast(user, "已上线")

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				s.BroadCast(user, "下线")
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read error:", err)
			}

			// 提取用户消息
			msg := string(buf)[:n-1]

			// 将得到消息广播
			s.BroadCast(user, msg)
		}
	}()

	// 当前handler阻塞
	select {

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