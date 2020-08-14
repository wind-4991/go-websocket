package impl

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	wsConn      *websocket.Conn
	inChan      chan []byte
	outChan     chan []byte
	closeChan   chan []byte
	isCloseChan bool
	mu          *sync.Mutex
}

// 初始化连接
func InitConnection(wsConn *websocket.Conn) (*Connection, error) {
	conn := &Connection{
		wsConn:      wsConn,
		inChan:      make(chan []byte, 1000),
		outChan:     make(chan []byte, 1000),
		closeChan:   make(chan []byte, 1),
		isCloseChan: false,
	}

	go conn.readLoop()
	go conn.writeLoop()

	return conn, nil
}

func (conn *Connection) ReadMsg() ([]byte, error) {
	var data []byte
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		return data, errors.New("connect is closed!")
	}

	return data, nil
}

func (conn *Connection) WriteMsg(data []byte) error {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		return errors.New("connect is closed!")
	}

	return nil
}

func (conn *Connection) Close() {
	if !conn.isCloseChan {
		conn.wsConn.Close()
		close(conn.closeChan)
		conn.isCloseChan = true
	}
}

// 读取消息存放到chan
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)

	// 循环读取
	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			goto ERR
		}

		select {
		case conn.inChan <- data:
		case <-conn.closeChan:
			goto ERR
		}
	}
ERR:
	conn.Close()
}

// 将chan取消息发送
func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)

	// 循环读取
	for {
		select {
		case data = <-conn.outChan:
			if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
				goto ERR
			}
		case <-conn.closeChan:
			goto ERR
		}
	}

ERR:
	conn.Close()
}
