package sockyts_test

import "testing"
import s "sockyts"

type mockAyxReader struct{}

func (r *mockAyxReader) ValuePushed(value string) {
	println(value)
}

type mockAyxWriter struct{}

func (r *mockAyxWriter) Write(value string) {
	println(value)
}

func TestRegisterAyxReader(t *testing.T) {
	reader := &mockAyxReader{}
	server := s.NewServer()
	server.RegisterAyxReader(`test`, reader)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegister2AyxReaderEndpoints(t *testing.T) {
	reader1 := &mockAyxReader{}
	reader2 := &mockAyxReader{}
	server := s.NewServer()
	server.RegisterAyxReader(`test1`, reader1)
	server.RegisterAyxReader(`test2`, reader2)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 2 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}

func TestRegisterAyxWriter(t *testing.T) {
	writer := &mockAyxWriter{}
	server := s.NewServer()
	server.RegisterAyxWriter(`test`, writer)
	endPoints := server.EndpointNames()
	if count := len(endPoints); count != 1 {
		t.Fatalf(`expected 1 endpoint but got %v`, count)
	}
	t.Logf(`endpoints: %v`, endPoints)
}
