package pubsub

import (
	"context"
)

type Marshaller func(interface{}) ([]byte, error)
type UnMarshaller func(source []byte, destination interface{}) error

const (
	NoMarshaller   = "Marshaller not specified"
	NoUnMarshaller = "UnMarshaller not specified"
)

type Publisher interface {
	Marshaller(marshaller Marshaller)
	Publish(context.Context, string, interface{}) error
	Errors() <-chan error
}

type Subscriber interface {
	UnMarshaller(unmarshaller UnMarshaller)
	Start() <-chan Message
	Errors() <-chan error
	Stop() error
}

type Message interface {
	UnMarshal(interface{}) error
	Done() error
}
