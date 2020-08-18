package client

import "net/http"

type Server interface {
	ConnectClient(addressName string, endpointName string) (<-chan string, chan<- string, error)
}

func SpinUpClient(server Server, w http.ResponseWriter, r *http.Request) {

}
