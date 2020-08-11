package sockyts

import (
	"fmt"
	"sync"
)

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
		endpoints: map[string]*Endpoint{},
		locker:    &sync.Mutex{},
	}
}

type server struct {
	endpoints map[string]*Endpoint
	locker    *sync.Mutex
}

func (s *server) RegisterAyxReader(endpointName string) <-chan string {
	channel := make(chan string)
	endpoint := s.registerEndpoint(endpointName)
	endpoint.AyxReaders = append(endpoint.AyxReaders, channel)
	return channel
}

func (s *server) RegisterAyxWriter(endpointName string) chan<- string {
	writer := &AyxWriter{
		WriteChan: make(chan string),
	}
	endpoint := s.registerEndpoint(endpointName)
	endpoint.AyxWriters = append(endpoint.AyxWriters, writer)
	return writer.WriteChan
}

func (s *server) registerEndpoint(endpointName string) *Endpoint {
	s.locker.Lock()
	endpoint, ok := s.endpoints[endpointName]
	if !ok {
		endpoint = &Endpoint{
			AyxWriteChan: make(chan string),
			AyxReadChan:  make(chan string),
			Clients:      make(map[*SockytClient]bool),
			locker:       &sync.Mutex{},
		}
		s.endpoints[endpointName] = endpoint
	}
	s.locker.Unlock()
	return endpoint
}

func (s *server) EndpointNames() []string {
	endpointNames := make([]string, len(s.endpoints))
	i := 0
	for key := range s.endpoints {
		endpointNames[i] = key
		i++
	}
	return endpointNames
}

func (s *server) ConnectClient(endpointName string) (<-chan string, chan<- string, error) {
	s.locker.Lock()
	endpoint, ok := s.endpoints[endpointName]
	s.locker.Unlock()
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
	for _, endpoint := range s.endpoints {
		for _, writer := range endpoint.AyxWriters {
			go s.forwardAyxWriter(endpoint, writer)
		}

		go s.writeToClientLoop(endpoint)
		go s.readFromClientLoop(endpoint)
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
	close(client.ReadChan)
	delete(endpoint.Clients, client)
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
