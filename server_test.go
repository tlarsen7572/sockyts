package sockyts_test

import "testing"
import s "sockyts"

func TestRegisterAyxReader(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxReader(`test`)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegister2AyxReaderEndpoints(t *testing.T) {
	server := s.NewServer()
	_ = server.RegisterAyxReader(`test1`)
	_ = server.RegisterAyxReader(`test2`)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 2 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegisterAyxWriter(t *testing.T) {
	server := s.NewServer()
	_, _ = server.RegisterAyxWriter(`test`)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}
