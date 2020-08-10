package sockyts

import (
	"fmt"
	"sync"
)

type Endpoint struct {
	AyxReaders   []chan string
	AyxWriters   []*AyxWriter
	AyxWriteChan chan string
	Clients      []*SockytClient
}

type AyxWriter struct {
	WriteChan chan string
}

type SockytClient struct {
	ReadChan  chan string
	WriteChan chan string
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
	endpoint.Clients = append(endpoint.Clients, client)
	return client.ReadChan, client.WriteChan, nil
}

func (s *server) Start() {
	for _, endpoint := range s.endpoints {
		for _, writer := range endpoint.AyxWriters {
			go func(w *AyxWriter) {
				for msg := range w.WriteChan {
					endpoint.AyxWriteChan <- msg
				}
			}(writer)
		}

		go func(e *Endpoint) {
			for msg := range e.AyxWriteChan {
				for _, clientReader := range e.Clients {
					clientReader.ReadChan <- msg
				}
			}
		}(endpoint)
	}
}
