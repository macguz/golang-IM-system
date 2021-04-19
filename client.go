package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

// NewClient 新建客户端
func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       99,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}

	client.conn = conn

	return client
}

// DealResponse 处理server响应，输出到屏幕上
func (c *Client) DealResponse() {
	io.Copy(os.Stdout, c.conn)
}

func (c *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		c.flag = flag
		return true
	} else {
		fmt.Println(">>>>请输入合法选项<<<<")
		return false
	}
}

// 查询在线用户
func (c *Client) SelectUser() {
	sendMsg := "who\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write error:", err)
		return
	}
}

func (c *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	c.SelectUser()
	fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>请输入消息内容, exit退出:")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {

			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n\n"

				_, err := c.conn.Write([]byte(sendMsg))

				if err != nil {
					fmt.Println("conn Write error:", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>请输入聊天内容, exit退出")
			fmt.Scanln(&chatMsg)
		}

		c.SelectUser()
		fmt.Println(">>>>请输入聊天对象[用户名], exit退出:")
		fmt.Scanln(&remoteName)
	}
}

func (c *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string

	fmt.Println(">>>>请输入聊天内容, exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"

			_, err := c.conn.Write([]byte(sendMsg))

			if err != nil {
				fmt.Println("conn Write error:", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>请输入聊天内容, exit退出")
		fmt.Scanln(&chatMsg)
	}
}

func (c *Client) UpdateName() bool {
	fmt.Println(">>>>请输入用户名:")
	fmt.Scanln(&c.Name)

	sendMsg := "rename|" + c.Name + "\n"
	_, err := c.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write error:", err)
		return false
	}
	return true
}

func (c *Client) Run() {
	for c.flag != 0 {
		for c.menu() != true {
		}

		switch c.flag {
		case 1:
			// 公聊模式
			c.PublicChat()
		case 2:
			// 私聊模式
			c.PrivateChat()
		case 3:
			// 更新用户名
			c.UpdateName()
			break
		}
	}
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口号(默认8888)")
}

func main() {

	// 解析命令行
	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>> 连接服务器失败...")
		return
	}

	// 单独开go程处理server响应
	go client.DealResponse()

	fmt.Println(">>>>> 连接服务器成功...")

	client.Run()
}
