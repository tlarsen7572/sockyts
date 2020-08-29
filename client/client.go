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

	fromServer, _, err := server.ConnectClient(r.URL.Host, r.URL.Path)
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	go func() {
		msg := <-fromServer
		writer, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		writer.Write([]byte(msg))
		err = writer.Close()
		if err != nil {
			return
		}
	}()
}
