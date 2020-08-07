package sockyts

type Server interface {
	RegisterAyxReader(endpointName string, reader AyxReader)
	EndpointNames() []string
}

type AyxReader interface {
	ValuePushed(string)
}
