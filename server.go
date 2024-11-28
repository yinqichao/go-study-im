package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip            string
	Port          int
	OnlineMap     map[string]*User
	onlineMapLock sync.RWMutex
	Message       chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (this_ *Server) ListenMessage() {
	for {
		message := <-this_.Message

		this_.onlineMapLock.RLock()
		for _, user := range this_.OnlineMap {
			user.C <- message
		}
		this_.onlineMapLock.RUnlock()
	}
}

func (this_ *Server) Handler(conn net.Conn) {
	user := NewUser(conn, this_)

	user.Online()

	isLive := make(chan bool)

	go func() {
		for {
			buf := make([]byte, 4096)
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("conn.Read err:", err)
				return
			}

			user.DoMessage(string(buf[:n-1]))

			isLive <- true
		}
	}()

	for {
		select {
		case <-isLive:
		case <-time.After(time.Minute * 5):
			user.sendMessage("静默超时，你被强踢了")
			close(user.C)
			return
		}
	}
}

func (this_ *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this_.Ip, this_.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	defer func(listener net.Listener) {
		_ = listener.Close()
	}(listener)

	go this_.ListenMessage()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener.Accept err:", err)
			continue
		}
		go this_.Handler(conn)
	}
}

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址")
	flag.IntVar(&serverPort, "port", 8080, "设置服务器端口")
}

func main() {
	flag.Parse()

	server := NewServer(serverIp, serverPort)
	server.Start()
}
