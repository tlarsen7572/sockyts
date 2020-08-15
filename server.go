package sockyts

import (
	"fmt"
	"sync"
)

type Address struct {
	Endpoints map[string]*Endpoint
	locker    *sync.Mutex
}

type Endpoint struct {
	AyxReaders   []chan string
	AyxWriters   []*AyxWriter
	AyxWriteChan chan string
	AyxReadChan  chan string
	Clients      map[*SockytClient]bool
	locker       *sync.Mutex
}

type AyxWriter struct {
	WriteChan chan string
}

type SockytClient struct {
	ReadChan  chan string
	WriteChan chan string
	index     int
}

func NewServer() Server {
	return &server{
		addresses: map[string]*Address{},
		locker:    &sync.Mutex{},
	}
}

type server struct {
	addresses map[string]*Address
	locker    *sync.Mutex
}

func (s *server) RegisterAyxReader(addressName string, endpointName string) <-chan string {
	channel := make(chan string)
	address := s.registerAddress(addressName)
	endpoint := address.registerEndpoint(endpointName)
	endpoint.AyxReaders = append(endpoint.AyxReaders, channel)
	return channel
}

func (s *server) RegisterAyxWriter(addressName string, endpointName string) chan<- string {
	writer := &AyxWriter{
		WriteChan: make(chan string),
	}
	address := s.registerAddress(addressName)
	endpoint := address.registerEndpoint(endpointName)
	endpoint.AyxWriters = append(endpoint.AyxWriters, writer)
	return writer.WriteChan
}

func (s *server) registerAddress(addressName string) *Address {
	s.locker.Lock()
	address, ok := s.addresses[addressName]
	if !ok {
		address = &Address{
			Endpoints: make(map[string]*Endpoint),
			locker:    &sync.Mutex{},
		}
		s.addresses[addressName] = address
	}
	s.locker.Unlock()
	return address
}

func (a *Address) registerEndpoint(endpointName string) *Endpoint {
	a.locker.Lock()
	endpoint, ok := a.Endpoints[endpointName]
	if !ok {
		endpoint = &Endpoint{
			AyxWriteChan: make(chan string),
			AyxReadChan:  make(chan string),
			Clients:      make(map[*SockytClient]bool),
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

	client := &SockytClient{
		ReadChan:  make(chan string),
		WriteChan: make(chan string),
	}
	endpoint.Clients[client] = true

	go s.forwardClientWriter(endpoint, client)
	return client.ReadChan, client.WriteChan, nil
}

func (s *server) Start() {
	for _, address := range s.addresses {
		for _, endpoint := range address.Endpoints {
			for _, writer := range endpoint.AyxWriters {
				go s.forwardAyxWriter(endpoint, writer)
			}

			go s.writeToClientLoop(endpoint)
			go s.readFromClientLoop(endpoint)
		}
	}
}

func (s *server) forwardAyxWriter(endpoint *Endpoint, writer *AyxWriter) {
	for msg := range writer.WriteChan {
		endpoint.AyxWriteChan <- msg
	}
}

func (s *server) forwardClientWriter(endpoint *Endpoint, client *SockytClient) {
	for msg := range client.WriteChan {
		endpoint.AyxReadChan <- msg
	}
	endpoint.locker.Lock()
	delete(endpoint.Clients, client)
	close(client.ReadChan)
	endpoint.locker.Unlock()
}

func (s *server) writeToClientLoop(endpoint *Endpoint) {
	for msg := range endpoint.AyxWriteChan {
		for clientReader := range endpoint.Clients {
			clientReader.ReadChan <- msg
		}
	}
}

func (s *server) readFromClientLoop(endpoint *Endpoint) {
	for msg := range endpoint.AyxReadChan {
		for _, ayxReader := range endpoint.AyxReaders {
			ayxReader <- msg
		}
	}
}
