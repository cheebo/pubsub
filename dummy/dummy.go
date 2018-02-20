package dummy

import (
	"context"
	"github.com/cheebo/pubsub"
)

type publisher struct{}

type subsriber struct {
	ch chan pubsub.Message
}

func NewPublisher() pubsub.Publisher {
	return publisher{}
}

func NewSubscriber(ch chan pubsub.Message) pubsub.Subscriber {
	return subsriber{
		ch: ch,
	}
}

func (p publisher) Publish(context.Context, string, interface{}) error {
	return nil
}

func (p publisher) Errors() <-chan error {
	return make(chan error)
}

func (s subsriber) Start() <-chan pubsub.Message {
	return s.ch
}

func (s subsriber) Errors() <-chan error {
	return make(chan error)
}

func (s subsriber) Stop() error {
	return nil
}
