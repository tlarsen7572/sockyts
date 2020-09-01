package client_test

import (
	"github.com/gorilla/websocket"
	"github.com/tlarsen7572/sockyts/client"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

type server struct {
	send         chan string
	receive      chan string
	msgsReceived chan string
}

func (s *server) ConnectClient(addressName string, endpointName string) (<-chan string, chan<- string, error) {
	s.send = make(chan string)
	s.receive = make(chan string)
	s.msgsReceived = make(chan string)
	go func() {
		for {
			msg, ok := <-s.receive
			if !ok {
				break
			}
			s.msgsReceived <- msg
		}
	}()
	return s.send, s.receive, nil
}

func TestConnectWithoutWebsocketProtocol(t *testing.T) {
	server := &server{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client.SpinUpClient(server, w, r)
	}))
	defer mockServer.Close()

	response, err := http.Get(mockServer.URL)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())

	}
	if response.StatusCode != 400 {
		t.Fatalf(`expected response 400 but got %v`, response.StatusCode)
	}
	t.Logf(`%v`, string(body))
}

func TestConnectAndReceiveMsg(t *testing.T) {
	server := &server{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client.SpinUpClient(server, w, r)
	}))
	defer mockServer.Close()

	u := url.URL{Scheme: "ws", Host: mockServer.URL[7:], Path: ``}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	defer conn.Close()

	server.send <- `hello world`
	okChannel := make(chan bool)
	go func() {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf(`expected no error but got: %v`, err.Error())
		}
		msgStr := string(msg)
		if msgStr != `hello world` {
			t.Fatalf(`expected 'hello world' but got '%v'`, msgStr)
		}
		t.Logf(`got message type %v with message '%v'`, msgType, msgStr)
		close(okChannel)
	}()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	select {
	case <-okChannel:
		return
	case <-ticker.C:
		t.Fatalf(`test timed out after 1 second`)
	}
}

func TestConnectAndReceiveMultipleMsgs(t *testing.T) {
	server := &server{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client.SpinUpClient(server, w, r)
	}))
	defer mockServer.Close()

	u := url.URL{Scheme: "ws", Host: mockServer.URL[7:], Path: ``}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	defer conn.Close()

	okChannel := make(chan bool)
	go func() {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf(`expected no error but got: %v`, err.Error())
		}
		msgStr := string(msg)
		if msgStr != `a` {
			t.Fatalf(`expected 'a' but got '%v'`, msgStr)
		}
		t.Logf(`got message type %v with message '%v'`, msgType, msgStr)

		msgType, msg, err = conn.ReadMessage()
		if err != nil {
			t.Fatalf(`expected no error but got: %v`, err.Error())
		}
		msgStr = string(msg)
		if msgStr != `b` {
			t.Fatalf(`expected 'b' but got '%v'`, msgStr)
		}
		t.Logf(`got message type %v with message '%v'`, msgType, msgStr)
		close(okChannel)
	}()
	server.send <- `a`
	server.send <- `b`
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	select {
	case <-okChannel:
		return
	case <-ticker.C:
		t.Fatalf(`test timed out after 1 second`)
	}
}

func TestConnectAndSendMessage(t *testing.T) {
	server := &server{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client.SpinUpClient(server, w, r)
	}))
	defer mockServer.Close()

	u := url.URL{Scheme: "ws", Host: mockServer.URL[7:], Path: ``}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`hello world`))
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	select {
	case msg := <-server.msgsReceived:
		if msg != `hello world` {
			t.Fatalf(`expected 'hello world' but got '%v'`, msg)
		}
	case <-ticker.C:
		t.Fatalf(`test timed out after 1 second`)
	}
}

func TestConnectAndSendMultipleMessages(t *testing.T) {
	server := &server{}
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		client.SpinUpClient(server, w, r)
	}))
	defer mockServer.Close()

	u := url.URL{Scheme: "ws", Host: mockServer.URL[7:], Path: ``}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte(`a`))
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	err = conn.WriteMessage(websocket.TextMessage, []byte(`b`))
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	select {
	case msg := <-server.msgsReceived:
		if msg != `a` {
			t.Fatalf(`expected 'a' but got '%v'`, msg)
		}
	case <-ticker.C:
		t.Fatalf(`test timed out after 1 second`)
	}

	select {
	case msg := <-server.msgsReceived:
		if msg != `b` {
			t.Fatalf(`expected 'b' but got '%v'`, msg)
		}
	case <-ticker.C:
		t.Fatalf(`test timed out after 1 second`)
	}
}
