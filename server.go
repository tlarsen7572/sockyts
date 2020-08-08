package sockyts

import "sync"

type Endpoint struct {
	AyxReaders []AyxReader
	AyxWriters []AyxWriter
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

func (s *server) RegisterAyxReader(endpointName string, reader AyxReader) {
	endpoint := s.registerEndpoint(endpointName)
	endpoint.AyxReaders = append(endpoint.AyxReaders, reader)
}

func (s *server) RegisterAyxWriter(endpointName string, writer AyxWriter) {
	endpoint := s.registerEndpoint(endpointName)
	endpoint.AyxWriters = append(endpoint.AyxWriters, writer)
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
