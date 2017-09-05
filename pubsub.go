package pubsub

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type Publisher interface {
	Publish(context.Context, string, proto.Message) error
}

type Subscriber interface {
	Start() <-chan Message
	Errors() <-chan error
	Stop() error
}

type Message interface {
	Message() []byte
	Done() error
}