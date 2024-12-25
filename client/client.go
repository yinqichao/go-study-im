package main

import (
	"flag"
	"fmt"
	"go-study-im/global"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	opt        int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		opt:        -1,
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return nil
	}

	client.conn = conn

	return client
}

func (this_ *Client) menu() bool {
	var opt int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.查询在线用户")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&opt)
	if err != nil {
		fmt.Println("fmt.Scanln err:", err)
		return false
	}
	if opt >= 0 && opt <= 4 {
		this_.opt = opt
		return true
	} else {
		fmt.Println("请输入合法数字")
		return false
	}
}

func (this_ *Client) Rename() bool {
	fmt.Println("请输入新的用户名")

	_, err := fmt.Scanln(&this_.Name)
	if err != nil {
		fmt.Println("fmt.Scanln err:", err)
		return false
	}
	sendMsg := "rename|" + this_.Name + "\n"
	_, err = this_.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (this_ *Client) PrivateChat() {
	var toName string
	var toMsg string

	//列出所有在线用户
	this_.Who()
	fmt.Println("请输入私聊对象的用户名，exit退出")
	_, err := fmt.Scanln(&toName)
	if err != nil {
		fmt.Println("fmt.Scanln err:", err)
		return
	}
	for toName != "exit" {
		if len(toName) > 0 {

			fmt.Println("请输入私聊内容，exit退出")
			_, err = fmt.Scanln(&toMsg)
			if err != nil {
				fmt.Println("fmt.Scanln err:", err)
				return
			}

			for toMsg != "exit" {
				if len(toMsg) > 0 {
					sendMsg := "to|" + toName + "|" + toMsg + "\n"
					_, err = this_.conn.Write([]byte(sendMsg))
					if err != nil {
						fmt.Println("conn.Write err:", err)
						return
					}
				}
				fmt.Println("请输入私聊内容，exit退出")
				_, err := fmt.Scanln(&toMsg)
				if err != nil {
					fmt.Println("fmt.Scanln err:", err)
					return
				}
			}
		}
		//列出所有在线用户
		this_.Who()
		fmt.Println("请输入私聊对象的用户名，exit退出")
		_, err := fmt.Scanln(&toName)
		if err != nil {
			fmt.Println("fmt.Scanln err:", err)
			return
		}
	}
}

func (this_ *Client) PublicChat() {
	var toMsg string

	fmt.Println("请输入公聊内容，exit退出")
	_, err := fmt.Scanln(&toMsg)
	if err != nil {
		fmt.Println("fmt.Scanln err:", err)
		return
	}
	for toMsg != "exit" {
		if len(toMsg) > 0 {
			sendMsg := toMsg + "\n"
			_, err = this_.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn.Write err:", err)
			}
		}
		fmt.Println("请输入公聊内容，exit退出")
		_, err := fmt.Scanln(&toMsg)
		if err != nil {
			fmt.Println("fmt.Scanln err:", err)
			return
		}
	}
}

func (this_ *Client) Who() {
	sendMsg := "who\n"
	_, err := this_.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
	}
}

func (this_ *Client) DealResponse() {
	_, err := io.Copy(os.Stdout, this_.conn)
	if err != nil {
		fmt.Println("io.Copy err:", err)
		return
	}
}

func (this_ *Client) Run() {
	for this_.opt != 0 {
		for this_.menu() != true {
		}

		switch this_.opt {
		case 1:
			this_.PublicChat()
			break
		case 2:
			this_.PrivateChat()
			break
		case 3:
			this_.Rename()
			break
		case 4:
			this_.Who()
			break
		}
	}
}

func init() {
	flag.StringVar(&global.ServerIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&global.ServerPort, "port", 8080, "设置服务器端口")
}

func main() {
	flag.Parse()

	client := NewClient(global.ServerIp, global.ServerPort)
	if client == nil {
		fmt.Println("连接服务器失败...")
		return
	}

	go client.DealResponse()

	fmt.Println("连接服务器成功...")

	client.Run()
}
