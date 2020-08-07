package sockyts_test

import "testing"
import s "sockyts"

type mockAyxReader struct{}

func (r *mockAyxReader) ValuePushed(value string) {
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
}
