package client_test

import (
	"github.com/gorilla/websocket"
	"github.com/tlarsen7572/sockyts/client"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type server struct{}

func (s *server) ConnectClient(addressName string, endpointName string) (<-chan string, chan<- string, error) {
	return nil, nil, nil
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

func TestConnect(t *testing.T) {
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

}
