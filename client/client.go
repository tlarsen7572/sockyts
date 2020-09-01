package client

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Server interface {
	ConnectClient(addressName string, endpointName string) (<-chan string, chan<- string, error)
}

func SpinUpClient(server Server, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	fromServer, toServer, err := server.ConnectClient(r.URL.Host, r.URL.Path)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	go func() {
		for {
			msg, ok := <-fromServer
			if !ok {
				break
			}
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				break
			}
		}
	}()
	go func() {
		for {
			messageType, msg, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if messageType != websocket.TextMessage {
				break
			}
			msgStr := string(msg)
			toServer <- msgStr
		}
	}()
}
