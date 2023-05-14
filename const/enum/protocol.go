package enum

type Protocol int8

const (
	ProtocolHTTP Protocol = iota
	ProtocolGRPC
)

func (p Protocol) String() string {
	switch p {
	case ProtocolHTTP:
		return "http"
	case ProtocolGRPC:
		return "gRPC"
	}
	panic("unknown protocol")
}
