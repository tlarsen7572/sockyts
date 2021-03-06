package server

type Server interface {
	RegisterAyxReader(addressName string, endpointName string) <-chan string
	RegisterAyxWriter(addressName string, endpointName string) (chan<- string, <-chan bool)
	AddressNames() []string
	EndpointNames(addressName string) []string
	ConnectClient(addressName string, endpointName string) (<-chan string, chan<- string, error)
	Start()
	Shutdown()
}
