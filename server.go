package sockyts

import "sync"

type Endpoint struct {
	AyxReaders []chan string
	AyxWriters []*AyxWriter
}

type AyxWriter struct {
	Closer    chan bool
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

func (s *server) RegisterAyxWriter(endpointName string) (<-chan bool, chan<- string) {
	writer := &AyxWriter{
		Closer:    make(chan bool),
		WriteChan: make(chan string),
	}
	endpoint := s.registerEndpoint(endpointName)
	endpoint.AyxWriters = append(endpoint.AyxWriters, writer)
	return writer.Closer, writer.WriteChan
}

func (s *server) registerEndpoint(endpointName string) *Endpoint {
	s.locker.Lock()
	endpoint, ok := s.endpoints[endpointName]
	if !ok {
		endpoint = &Endpoint{}
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
