package main

import (
	"context"
	"net"
	"encoding/json"
)

var clientLog *Log
var lastContent []byte
var msgHandler map[ContentType]func(msg *TcpMsg) error
var isServer bool
var clipboardManager *ClipboardManager
var keyboardManager *KeyboardManager
var lastKeyboardState *KeyboardState

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
	}

	// clipboard
	clipboardManager = NewClipboardManager()
	err = clipboardManager.Init()
	if err != nil {
		clientLog.Log("init clipboard error: %s", err.Error())
		panic(err)
	}

	// keyboard
	keyboardManager = NewKeyboardManager()
	initialState, err := keyboardManager.GetState()
	if err != nil {
		clientLog.Log("init keyboard error: %s", err.Error())
		panic(err)
	}
	lastKeyboardState = initialState

	clipboardCh := clipboardManager.Watch(context.Background())
	keyboardCh := keyboardManager.Watch(context.Background())
	msgCh := tcp.Watch()
	handler(tcp, clipboardCh, keyboardCh, msgCh)
}

func handler(tcp *Tcp, clipboardCh <-chan []byte, keyboardCh <-chan *KeyboardState, msgCh <-chan *TcpMsg) {
	defer tcp.Close()
	lastContent = clipboardManager.Read()
	for {
		select {
		case content := <-clipboardCh:
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

		case state := <-keyboardCh:
			if state == nil || (state.NumLock == lastKeyboardState.NumLock && 
			   state.CapsLock == lastKeyboardState.CapsLock) {
				continue
			}
			stateBytes, err := json.Marshal(state)
			if err != nil {
				clientLog.Log("marshal keyboard state error: %s", err.Error())
				continue
			}
			err = tcp.Send(&TcpMsg{
				Content: stateBytes,
				Type:    CTKeyboardState,
			})
			if err != nil {
				clientLog.Log("send keyboard state error: %s", err.Error())
				continue
			}
			lastKeyboardState = state

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
	msgHandler[CTKeyboardState] = handlerKeyboardState
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

func handlerKeyboardState(msg *TcpMsg) error {
	var state KeyboardState
	if err := json.Unmarshal(msg.Content, &state); err != nil {
		return err
	}

	if lastKeyboardState != nil && 
	   state.NumLock == lastKeyboardState.NumLock && 
	   state.CapsLock == lastKeyboardState.CapsLock {
		return nil
	}

	err := keyboardManager.SetState(&state)
	if err != nil {
		return err
	}
	lastKeyboardState = &state
	return nil
}
