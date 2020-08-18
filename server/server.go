package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

type Status int

const (
	NotStarted Status = 0
	Running    Status = 1
	Closed     Status = 2
)

type serveAddress struct {
	Endpoints map[string]*serveEndpoint
	locker    *sync.Mutex
	mux       *http.ServeMux
	server    *http.Server
}

type serveEndpoint struct {
	AyxReaders   []chan string
	AyxWriters   []*ayxWriter
	AyxWriteChan chan string
	AyxReadChan  chan string
	Clients      map[*sockytClient]bool
	locker       *sync.Mutex
}

type ayxWriter struct {
	WriteChan chan string
	CloseChan chan bool
}

type sockytClient struct {
	ReadChan  chan string
	WriteChan chan string
	index     int
}

func NewServer() Server {
	return &server{
		addresses: map[string]*serveAddress{},
		locker:    &sync.Mutex{},
	}
}

type server struct {
	addresses map[string]*serveAddress
	locker    *sync.Mutex
	status    Status
}

func (s *server) RegisterAyxReader(addressName string, endpointName string) <-chan string {
	channel := make(chan string)
	address := s.registerAddress(addressName)
	endpoint := address.registerEndpoint(endpointName)
	endpoint.AyxReaders = append(endpoint.AyxReaders, channel)
	return channel
}

func (s *server) RegisterAyxWriter(addressName string, endpointName string) (chan<- string, <-chan bool) {
	writer := &ayxWriter{
		WriteChan: make(chan string),
		CloseChan: make(chan bool),
	}
	address := s.registerAddress(addressName)
	endpoint := address.registerEndpoint(endpointName)
	endpoint.AyxWriters = append(endpoint.AyxWriters, writer)
	return writer.WriteChan, writer.CloseChan
}

func (s *server) registerAddress(addressName string) *serveAddress {
	s.locker.Lock()
	address, ok := s.addresses[addressName]
	if !ok {
		address = &serveAddress{
			Endpoints: make(map[string]*serveEndpoint),
			locker:    &sync.Mutex{},
		}
		s.addresses[addressName] = address
	}
	s.locker.Unlock()
	return address
}

func (a *serveAddress) registerEndpoint(endpointName string) *serveEndpoint {
	a.locker.Lock()
	endpoint, ok := a.Endpoints[endpointName]
	if !ok {
		endpoint = &serveEndpoint{
			AyxWriteChan: make(chan string),
			AyxReadChan:  make(chan string),
			Clients:      make(map[*sockytClient]bool),
			locker:       &sync.Mutex{},
		}
		a.Endpoints[endpointName] = endpoint
	}
	a.locker.Unlock()
	return endpoint
}

func (s *server) EndpointNames(addressName string) []string {
	address, ok := s.addresses[addressName]
	if !ok {
		return nil
	}
	endpointNames := make([]string, len(address.Endpoints))
	i := 0
	for key := range address.Endpoints {
		endpointNames[i] = key
		i++
	}
	return endpointNames
}

func (s *server) AddressNames() []string {
	addressNames := make([]string, len(s.addresses))
	i := 0
	for key := range s.addresses {
		addressNames[i] = key
		i++
	}
	return addressNames
}

func (s *server) ConnectClient(addressName string, endpointName string) (<-chan string, chan<- string, error) {
	s.locker.Lock()
	address, ok := s.addresses[addressName]
	s.locker.Unlock()
	if !ok {
		return nil, nil, fmt.Errorf(`address %v is not valid`, addressName)
	}
	address.locker.Lock()
	endpoint, ok := address.Endpoints[endpointName]
	address.locker.Unlock()
	if !ok {
		return nil, nil, fmt.Errorf(`endpoint %v is not valid`, endpointName)
	}

	client := &sockytClient{
		ReadChan:  make(chan string),
		WriteChan: make(chan string),
	}

	s.locker.Lock()
	okToAdd := s.status == Running
	if !okToAdd {
		s.locker.Unlock()
		return nil, nil, fmt.Errorf(`server is not accepting clients`)
	}
	endpoint.Clients[client] = true
	s.locker.Unlock()

	go s.forwardClientWriter(endpoint, client)
	return client.ReadChan, client.WriteChan, nil
}

func (s *server) Start() {
	s.locker.Lock()
	defer s.locker.Unlock()

	if s.status > NotStarted {
		return
	}

	for addressName, address := range s.addresses {
		address.mux = http.NewServeMux()
		address.server = &http.Server{Addr: addressName, Handler: address.mux}
		address.mux.HandleFunc(addressName, func(w http.ResponseWriter, r *http.Request) {})
		for _, endpoint := range address.Endpoints {
			for _, writer := range endpoint.AyxWriters {
				go s.forwardAyxWriter(endpoint, writer)
			}

			go s.writeToClientLoop(endpoint)
			go s.readFromClientLoop(endpoint)
		}
		go func(a *serveAddress) {
			_ = a.server.ListenAndServe()
		}(address)
	}
	s.status = Running
}

func (s *server) forwardAyxWriter(endpoint *serveEndpoint, writer *ayxWriter) {
	for msg := range writer.WriteChan {
		endpoint.AyxWriteChan <- msg
	}
}

func (s *server) forwardClientWriter(endpoint *serveEndpoint, client *sockytClient) {
	for msg := range client.WriteChan {
		endpoint.AyxReadChan <- msg
	}
	endpoint.tryDisconnectClient(client)
}

func (endpoint *serveEndpoint) tryDisconnectClient(client *sockytClient) {
	endpoint.locker.Lock()
	_, ok := endpoint.Clients[client]
	if ok {
		delete(endpoint.Clients, client)
		close(client.ReadChan)
	}
	endpoint.locker.Unlock()
}

func (s *server) writeToClientLoop(endpoint *serveEndpoint) {
	for msg := range endpoint.AyxWriteChan {
		for clientReader := range endpoint.Clients {
			clientReader.ReadChan <- msg
		}
	}
	for clientReader := range endpoint.Clients {
		close(clientReader.ReadChan)
	}
}

func (s *server) readFromClientLoop(endpoint *serveEndpoint) {
	for msg := range endpoint.AyxReadChan {
		for _, ayxReader := range endpoint.AyxReaders {
			ayxReader <- msg
		}
	}
}

func (s *server) Shutdown() {
	s.locker.Lock()
	if s.status >= Closed {
		s.locker.Unlock()
		return
	}
	s.status = Closed
	s.locker.Unlock()

	for _, address := range s.addresses {
		for _, endpoint := range address.Endpoints {
			for client := range endpoint.Clients {
				endpoint.tryDisconnectClient(client)
			}
			for _, ayxWriter := range endpoint.AyxWriters {
				close(ayxWriter.CloseChan)
			}
		}
		_ = address.server.Shutdown(context.Background())
	}
}
