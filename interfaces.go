package sockyts

type Server interface {
	RegisterAyxReader(endpointName string) <-chan string
	RegisterAyxWriter(endpointName string) chan<- string
	EndpointNames() []string
	ConnectClient(endpointName string) (<-chan string, chan<- string, error)
	Start()
}
