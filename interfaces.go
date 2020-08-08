package sockyts

type Server interface {
	RegisterAyxReader(endpointName string) <-chan string
	RegisterAyxWriter(endpointName string) (<-chan bool, chan<- string)
	EndpointNames() []string
}
