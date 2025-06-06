package main

import (
	"encoding/json"
	"net"

	"github.com/google/uuid"
)

const headerLen int = 8

type ContentType int

const (
	CTUnknown ContentType = iota
	CTSystem
	CTPassword
	CTText
	CTImg
	CTFile
	CTKeyboardState
)

type SystemContent struct {
	Text string
	Code int
}

func (sc *SystemContent) Bytes() []byte {
	bytes, _ := json.Marshal(sc)
	return bytes
}

type TcpMsg struct {
	Name    string
	Content []byte
	Type    ContentType
	To      string `json:"-"`
}

type Tcp struct {
	role    RoleType
	conn    net.Conn
	id      string
	watchCh chan *TcpMsg
	log     *Log
}

func NewTcp(conn net.Conn, role RoleType, log *Log) *Tcp {
	return &Tcp{
		role: role,
		conn: conn,
		id:   conn.RemoteAddr().String() + "-" + uuid.New().String()[:5],
		log:  log,
	}
}

func (t *Tcp) Watch() <-chan *TcpMsg {
	if t.watchCh != nil {
		return t.watchCh
	}
	t.watchCh = make(chan *TcpMsg, 1)
	go func() {
		for {
			msg, err := t.Read()
			if err != nil {
				panic(err)
			}
			t.watchCh <- msg
		}
	}()
	return t.watchCh
}

func (t *Tcp) Read() (*TcpMsg, error) {
	return t.read()
}

func (t *Tcp) Send(msg *TcpMsg) error {
	return t.send(msg.Name, msg.Content, msg.Type)
}

// writeFull 确保完整发送指定数量的字节
func (t *Tcp) writeFull(data []byte) error {
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := t.conn.Write(data[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	return nil
}

func (t *Tcp) send(name string, contentBytes []byte, contentType ContentType) error {
	msg := &TcpMsg{
		Name:    name,
		Content: contentBytes,
		Type:    contentType,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	contentLen := len(msgBytes)
	contentLenBytes := Int64ToBytes(int64(contentLen))

	// 发送长度信息
	err = t.writeFull(contentLenBytes)
	if err != nil {
		return err
	}

	// 发送消息内容
	err = t.writeFull(msgBytes)
	if err != nil {
		return err
	}

	msg.To = t.conn.RemoteAddr().String()
	t.log.LogSendMsg(msg)
	return nil
}

// readFull 确保完整读取指定数量的字节
func (t *Tcp) readFull(buf []byte) error {
	bytesRead := 0
	for bytesRead < len(buf) {
		n, err := t.conn.Read(buf[bytesRead:])
		if err != nil {
			return err
		}
		bytesRead += n
	}
	return nil
}

func (t *Tcp) read() (*TcpMsg, error) {
	// read content len
	lenInfoBytes := make([]byte, headerLen)
	err := t.readFull(lenInfoBytes)
	if err != nil {
		return nil, err
	}

	msgLen := BytesToInt64(lenInfoBytes)
	// read content
	msgBytes := make([]byte, msgLen)
	err = t.readFull(msgBytes)
	if err != nil {
		return nil, err
	}

	msg := &TcpMsg{}
	err = json.Unmarshal(msgBytes, msg)
	if err != nil {
		return nil, err
	}
	msg.To = t.conn.RemoteAddr().String()
	t.log.LogReadMsg(msg)
	return msg, nil
}

func (t *Tcp) GetTcpID() string {
	return t.id
}

func (t *Tcp) Close() {
	if t.watchCh != nil {
		close(t.watchCh)
	}
	err := t.conn.Close()
	if err != nil {
		t.log.Log("close conn error: %s", err.Error())
	}
}
