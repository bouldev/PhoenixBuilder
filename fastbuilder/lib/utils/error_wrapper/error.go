package error_wrapper

import "fmt"

type ClientServerError[SE any] interface {
	// list of strings that send to client, should be in English and some of them can be translate
	// e.g. ["can not find %v(%v) specific", "rental server", "2401"]
	ClientSide() []string
	// detailed error info kept in server for debugging
	ServerSide() SE
	// standard error interface
	Error() string
}

type BasicClientServerError[SE any] struct {
	clientSide []string
	serverSide SE
}

func (b *BasicClientServerError[SE]) ClientSide() []string {
	return b.clientSide
}

func (b *BasicClientServerError[SE]) ServerSide() SE {
	return b.serverSide
}

func (b *BasicClientServerError[SE]) Error() string {
	return fmt.Sprintf("client side: %v, server side: %v", b.clientSide, b.serverSide)
}

func NewSEError[SE any](clientSide []string, serverSide SE) ClientServerError[SE] {
	return &BasicClientServerError[SE]{clientSide, serverSide}
}
