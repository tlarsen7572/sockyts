package sockyts_test

import (
	"strconv"
	"testing"
	"time"
)
import s "sockyts"

func TestRegisterAyxReader(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxReader(`address`, `test`)
	addresses := server.AddressNames()
	if count := len(addresses); count != 1 {
		t.Fatalf(`expected 1 address but got %v`, count)
	}
	endPoints := server.EndpointNames(`address`)
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegister2AyxReaderEndpoints(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxReader(`address`, `test1`)
	_ = server.RegisterAyxReader(`address`, `test2`)
	addresses := server.AddressNames()
	if count := len(addresses); count != 1 {
		t.Fatalf(`expected 1 address but got %v`, count)
	}
	endPoints := server.EndpointNames(`address`)
	if count := len(endPoints); count != 2 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegisterAyxWriter(t *testing.T) {
	server := s.NewServer()
	_, _ = server.RegisterAyxWriter(`address`, `test`)
	addresses := server.AddressNames()
	if count := len(addresses); count != 1 {
		t.Fatalf(`expected 1 address but got %v`, count)
	}
	endPoints := server.EndpointNames(`address`)
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestAddClientAndReadMsg(t *testing.T) {
	server := s.NewServer()
	writeChan, _ := server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	clientRead, _, err := server.ConnectClient(`address`, `test`)
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
	_, _ = server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	_, _, err := server.ConnectClient(`address`, `invalid`)
	if err == nil {
		t.Fatalf(`expected an error but got none`)
	}
}

func TestAddClientAndWriteMsg(t *testing.T) {
	server := s.NewServer()
	readChan := server.RegisterAyxReader(`address`, `test`)
	server.Start()
	_, clientWrite, err := server.ConnectClient(`address`, `test`)
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
	_, _ = server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	_, clientWrite, err := server.ConnectClient(`address`, `test`)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	clientWrite <- `hello world`
	t.Logf(`finished without deadlocks`)
}

func TestWriteWithNoClients(t *testing.T) {
	server := s.NewServer()
	writeChan, _ := server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	writeChan <- `hello world`
	t.Logf(`finished without deadlocks`)
}

func TestClosingClientWriteChannelRemovesClientFromEndpoint(t *testing.T) {
	server := s.NewServer()
	writeChan, _ := server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	clientRead, clientWrite, _ := server.ConnectClient(`address`, `test`)
	close(clientWrite)
	_, ok := <-clientRead
	if ok {
		t.Fatalf(`clientRead was not closed but it should have been`)
	}
	writeChan <- `hello world`
	time.Sleep(300 * time.Millisecond)
	t.Logf(`no panic, we didn't send 'hello world' to the closed reader`)
}

func TestShutdownServer(t *testing.T) {
	server := s.NewServer()
	ayxWriter, ayxCloser := server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	clientReader, _, _ := server.ConnectClient(`address`, `test`)
	server.Shutdown()
	ayxWriter <- `hello world`
	_, ok := <-ayxCloser
	if ok {
		t.Fatalf(`expected the ayxCloser channel to close but it did not`)
	}
	msg, ok := <-clientReader
	if ok {
		t.Logf(`msg received by client: %v`, msg)
		t.Fatalf(`expected the clientReader channel to close but it did not`)
	}
}

func TestAddClientAfterShutdown(t *testing.T) {
	server := s.NewServer()
	_, _ = server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	server.Shutdown()
	_, _, err := server.ConnectClient(`address`, `test`)
	if err == nil {
		t.Fatalf(`expected an error but got non`)
	}
	t.Logf(`error %v`, err.Error())
}

func TestSpinUpFor1Second(t *testing.T) {
	server := s.NewServer()
	ayxWriter, ayxCloser := server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	go func() {
		innerSleepChan := make(chan bool)
		i := 0
		for {
			go func() {
				time.Sleep(10 * time.Millisecond)
				innerSleepChan <- true
			}()
			select {
			case <-ayxCloser:
				close(ayxWriter)
				return
			case <-innerSleepChan:
				ayxWriter <- strconv.Itoa(i)
				i++
			}
		}
	}()
	clientRead, _, err := server.ConnectClient(`address`, `test`)
	if err != nil {
		t.Fatalf(`expected no error but got: %v`, err.Error())
	}
	outerSleepChan := make(chan string)
	go func() {
		time.Sleep(1 * time.Second)
		close(outerSleepChan)
	}()
	for {
		select {
		case <-outerSleepChan:
			server.Shutdown()
			return
		case msg := <-clientRead:
			t.Logf(msg)
		}
	}
}

func TestShutdownShouldBeIdempotent(t *testing.T) {
	server := s.NewServer()
	_, _ = server.RegisterAyxWriter(`address`, `test`)
	server.Start()
	server.Shutdown()
	server.Shutdown()
	t.Logf(`no panic happened because closed channels were closed twice, shutdown is idempotent`)
}
