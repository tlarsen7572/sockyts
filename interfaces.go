package sockyts

type Server interface {
	RegisterAyxReader(endpointName string, reader AyxReader)
	RegisterAyxWriter(endpointName string, writer AyxWriter)
	EndpointNames() []string
}

type AyxReader interface {
	ValuePushed(string)
}

type AyxWriter interface {
	Write(string)
}
