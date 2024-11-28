package main

import (
	"fmt"
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

func NewUser(conn net.Conn, server *Server) *User {
	remoteAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   remoteAddr,
		Addr:   remoteAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.Listen()

	return user
}

func (this_ *User) Listen() {
	//只有C chan关闭了，range才会推出
	for msg := range this_.C {
		this_.sendMessage(msg)
	}
	err := this_.conn.Close()
	if err != nil {
		panic(err)
	}
}

func (this_ *User) Online() {
	this_.server.onlineMapLock.Lock()
	this_.server.OnlineMap[this_.Name] = this_
	this_.server.onlineMapLock.Unlock()

	this_.broadMessage("已上线")
}

func (this_ *User) Offline() {
	this_.server.onlineMapLock.Lock()
	delete(this_.server.OnlineMap, this_.Name)
	this_.server.onlineMapLock.Unlock()

	this_.broadMessage("下线")
}

func (this_ *User) DoMessage(msg string) {
	//查询全部用户
	if msg == "who" {
		this_.server.onlineMapLock.RLock()
		for _, user := range this_.server.OnlineMap {
			message := "[" + user.Addr + "]" + user.Name + ":在线..."
			this_.sendMessage(message)
		}
		this_.server.onlineMapLock.RUnlock()
		return
	}

	//修改用户名
	if len(msg) > 7 && msg[:7] == "rename|" && strings.Count(msg, "|") == 1 {
		newName := strings.Split(msg, "|")[1]
		if strings.TrimSpace(newName) == "" {
			this_.sendMessage("用户名不能为空")
			return
		}
		_, isExist := this_.server.OnlineMap[newName]
		if isExist {
			this_.sendMessage("该用户名已存在")
			return
		}
		this_.server.onlineMapLock.Lock()
		delete(this_.server.OnlineMap, this_.Name)
		this_.server.OnlineMap[newName] = this_
		this_.server.onlineMapLock.Unlock()

		this_.Name = newName
		this_.sendMessage("用户名成功修改为:" + newName)
		return
	}

	//私聊
	if len(msg) > 3 && msg[:3] == "to|" && strings.Count(msg, "|") == 2 {
		toName := strings.Split(msg, "|")[1]
		toMsg := strings.Split(msg, "|")[2]
		toUser, isExist := this_.server.OnlineMap[toName]
		if !isExist {
			this_.sendMessage("用户名[" + toName + "]不存在")
			return
		}
		if toName == this_.Name {
			this_.sendMessage("不能将消息发给自己")
			return
		}
		if strings.TrimSpace(toMsg) == "" {
			this_.sendMessage("消息不能为空")
			return
		}
		toUser.sendMessage("用户名[" + this_.Name + "]发来了一条消息:" + toMsg)
		return
	}

	//广播
	this_.broadMessage(msg)
}

func (this_ *User) sendMessage(msg string) {
	_, err := this_.conn.Write([]byte(msg + "\n"))
	if err != nil {
		fmt.Println("conn.Write err:", err)
	}
}

func (this_ *User) broadMessage(msg string) {
	message := "[" + this_.Addr + "]" + this_.Name + ":" + msg
	this_.server.Message <- message
}
