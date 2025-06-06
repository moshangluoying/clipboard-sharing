package main

import (
	"context"
	"net"
)

var clientLog *Log
var lastContent []byte
var msgHandler map[ContentType]func(msg *TcpMsg) error
var isServer bool
var clipboardManager *ClipboardManager

func runClient() {
	// init log
	clientLog = NewLog(RTClient)
	// init msg Handler
	initMsgHandler()
	// connect server
	if config.Role == RTServer {
		config.Host = "127.0.0.1"
	}
	conn, err := net.Dial("tcp", config.Host+":"+config.Port)
	if err != nil {
		clientLog.Log("connect server tcp error: %s", err.Error())
		panic(err)
	}
	tcp := NewTcp(conn, RTClient, clientLog)
	// send password
	err = tcp.Send(&TcpMsg{
		Name:    "",
		Content: []byte(config.Password),
		Type:    CTPassword,
	})
	if err != nil {
		clientLog.Log("send password error: %s", err.Error())
		panic(err)
	}	// clipboard
	clipboardManager = NewClipboardManager()
	err = clipboardManager.Init()
	if err != nil {
		clientLog.Log("init clipboard error: %s", err.Error())
		panic(err)
	}

	clipboardCh := clipboardManager.Watch(context.Background())
	msgCh := tcp.Watch()
	clipboardHandler(tcp, clipboardCh, msgCh)
}

func clipboardHandler(tcp *Tcp, clipboardCh <-chan []byte, msgCh <-chan *TcpMsg) {
	defer tcp.Close()
	lastContent = clipboardManager.Read()
	for {
		var content []byte
		select {
		case content = <-clipboardCh:
			if string(content) == string(lastContent) {
				continue
			}
			err := tcp.Send(&TcpMsg{
				Content: content,
				Type:    CTText,
			})
			if err != nil {
				clientLog.Log("send msg error: %s", err.Error())
				panic(err)
			}
			lastContent = content
		case msg := <-msgCh:
			f, ok := msgHandler[msg.Type]
			if !ok {
				continue
			}
			err := f(msg)
			if err != nil {
				clientLog.Log("handler msg error: %s", err.Error())
				panic(err)
			}
		}
	}
}

func initMsgHandler() {
	msgHandler = make(map[ContentType]func(msg *TcpMsg) error)
	msgHandler[CTText] = handlerText
}

func handlerText(msg *TcpMsg) error {
	if string(msg.Content) == string(lastContent) {
		return nil
	}
	err := clipboardManager.Write(msg.Content)
	if err != nil {
		return err
	}
	lastContent = msg.Content
	return nil
}
