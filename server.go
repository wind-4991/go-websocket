package main

import (
	"github.com/gorilla/websocket"
	"go-websocket/impl"
	"log"
	"net/http"
	"time"
)

var(
	upgrader = &websocket.Upgrader{
		// 运行跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var(
		conn *websocket.Conn
		err error
		data, returnData []byte
		wsConn *impl.Connection
	)
	// 升级连接为ws
	if conn, err = upgrader.Upgrade(w, r, nil); err != nil {
		log.Println("upgrade error:" + err.Error())
		return
	}

	if wsConn, err = impl.InitConnection(conn); err != nil {
		log.Println("upgrade error:" + err.Error())
		return
	}

	defer wsConn.Close()

	// 发送心跳
	go func() {
		for {
			if err := wsConn.WriteMsg([]byte("heart message...")); err != nil {
				log.Println(err.Error())
				return
			}

			// 每3秒执行一次
			time.Sleep(3 * time.Second)
		}
	}()

	for {
		// 通过连接获取客户端发送的消息
		if data, err = wsConn.ReadMsg(); err != nil {
			log.Println(err.Error())
			break
		}

		log.Println("客户端发送的消息："+string(data))
		log.Println(conn.RemoteAddr().String())
		log.Println(conn.RemoteAddr().Network())

		// 返回客户端发送的消息给客户端
		returnData = data
		if err = wsConn.WriteMsg(returnData); err != nil {
			log.Println(err.Error())
			break
		}
	}

	return
}

func main() {
	http.HandleFunc("/ws", wsHandler)
	servErr := http.ListenAndServe("0.0.0.0:8000", nil)
	if servErr != nil {
		log.Println(servErr)
	}
}
