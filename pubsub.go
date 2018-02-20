package pubsub

import (
	"context"
	"errors"
)

type Marshaller func(interface{}) ([]byte, error)
type UnMarshaller func(source []byte, destination interface{}) error

const (
	NoMarshaller   = "marshaller not specified"
	NoUnMarshaller = "unmarshaller not specified"
)

var ErrorNoMarshaller = errors.New(NoMarshaller)
var ErrorNoUnMarshaller = errors.New(NoUnMarshaller)

type Publisher interface {
	Publish(context.Context, string, interface{}) error
	Errors() <-chan error
}

type Subscriber interface {
	Start() <-chan Message
	Errors() <-chan error
	Stop() error
}

type Message interface {
	UnMarshal(interface{}) error
	Done() error
}
