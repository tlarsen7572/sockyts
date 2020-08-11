package sockyts_test

import (
	"testing"
	"time"
)
import s "sockyts"

func TestRegisterAyxReader(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxReader(`test`)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegister2AyxReaderEndpoints(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxReader(`test1`)
	_ = server.RegisterAyxReader(`test2`)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 2 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegisterAyxWriter(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxWriter(`test`)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestAddClientAndReadMsg(t *testing.T) {
	server := s.NewServer()
	writeChan := server.RegisterAyxWriter(`test`)
	server.Start()
	clientRead, _, err := server.ConnectClient(`test`)
	if err != nil {
		t.Fatalf(`expected no error but got %v`, err.Error())
	}
	writeChan <- `hello world`
	msg := <-clientRead
	if msg != `hello world` {
		t.Fatalf(`expected 'hello world' but got '%v'`, msg)
	}
}

func TestAddClientToInvalidEndpoint(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxWriter(`test`)
	server.Start()
	_, _, err := server.ConnectClient(`invalid`)
	if err == nil {
		t.Fatalf(`expected an error but got none`)
	}
}

func TestAddClientAndWriteMsg(t *testing.T) {
	server := s.NewServer()
	readChan := server.RegisterAyxReader(`test`)
	server.Start()
	_, clientWrite, err := server.ConnectClient(`test`)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	clientWrite <- `hello world`
	msg := <-readChan
	if msg != `hello world` {
		t.Fatalf(`expected 'hello world' but got '%v'`, msg)
	}
}

func TestAddClientAndWriteMsgNoAyxReaders(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxWriter(`test`)
	server.Start()
	_, clientWrite, err := server.ConnectClient(`test`)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	clientWrite <- `hello world`
	t.Logf(`finished without deadlocks`)
}

func TestWriteWithNoClients(t *testing.T) {
	server := s.NewServer()
	writeChan := server.RegisterAyxWriter(`test`)
	server.Start()
	writeChan <- `hello world`
	t.Logf(`finished without deadlocks`)
}

func TestClosingClientWriteChannelRemovesClientFromEndpoint(t *testing.T) {
	server := s.NewServer()
	writeChan := server.RegisterAyxWriter(`test`)
	server.Start()
	clientRead, clientWrite, _ := server.ConnectClient(`test`)
	close(clientWrite)
	_, ok := <-clientRead
	if ok {
		t.Fatalf(`clientRead was not closed but it should have been`)
	}
	writeChan <- `hello world`
	time.Sleep(300 * time.Millisecond)
	t.Logf(`no panic, we didn't send 'hello world' to the closed reader`)
}
