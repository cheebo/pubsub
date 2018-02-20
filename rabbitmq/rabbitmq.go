package rabbitmq

import (
	"context"
	"github.com/assembla/cony"
	"github.com/cheebo/go-config/types"
	"github.com/cheebo/pubsub"
	"github.com/streadway/amqp"
	"time"
)

type publisher struct {
	config     *types.AMQPConfig
	client     *cony.Client
	publisher  *cony.Publisher
	errors     chan error
	marshaller pubsub.Marshaller
}

type subscriber struct {
	unmarshaller pubsub.UnMarshaller
	config       *types.AMQPConfig
	client       *cony.Client
	consumer     *cony.Consumer
	stop         chan chan error
	errors       chan error
}

type message struct {
	unmarshaller pubsub.UnMarshaller
	body         []byte
	ack          func(bool) error
	nack         func(bool, bool) error
}

func NewPublisher(cfg *types.AMQPConfig, marshaller pubsub.Marshaller) (pubsub.Publisher, error) {
	pub := &publisher{
		config:     cfg,
		marshaller: marshaller,
	}

	pub.client = cony.NewClient(
		cony.URL(cfg.URL),
		cony.Backoff(cony.DefaultBackoff),
	)

	// Declare the exchange we'll be using
	exc := cony.Exchange{
		Name:       cfg.Exchange,
		Kind:       cfg.Kind,
		AutoDelete: cfg.AutoDelete,
		Durable:    cfg.Durable,
	}
	pub.client.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	// Declare and register a publisher
	pub.publisher = cony.NewPublisher(exc.Name, cfg.Key)
	pub.client.Publish(pub.publisher)

	pub.errors = make(chan error, 100)

	go func(cli *cony.Client, errors chan error) {
		for cli.Loop() {
			select {
			case err := <-cli.Errors():
				errors <- err
				//case <-cli.Blocking():
				//
			}
		}
	}(pub.client, pub.errors)

	return pub, nil
}

func (p *publisher) Publish(ctx context.Context, key string, msg interface{}) error {
	if p.marshaller == nil {
		return pubsub.ErrorNoMarshaller
	}

	body, err := p.marshaller(msg)
	if err != nil {
		return err
	}

	publishing := amqp.Publishing{
		DeliveryMode: uint8(p.config.DeliveryMode),
		ContentType:  "application/x-protobuf",
		Timestamp:    time.Now(),
		Body:         body,
	}

	if len(key) > 0 {
		return p.publisher.PublishWithRoutingKey(publishing, key)
	} else {
		return p.publisher.Publish(publishing)
	}
}

func (p *publisher) Errors() <-chan error {
	return p.errors
}

func NewSubscriber(cfg *types.AMQPConfig, unmarshaller pubsub.UnMarshaller) (pubsub.Subscriber, error) {
	client := cony.NewClient(
		cony.URL(cfg.URL),
		cony.Backoff(cony.DefaultBackoff),
	)

	// Declarations
	// The queue name will be supplied by the AMQP server
	que := &cony.Queue{
		Name:       cfg.Queue,
		AutoDelete: cfg.AutoDelete,
		Durable:    cfg.Durable,
	}
	exc := cony.Exchange{
		Name:       cfg.Exchange,
		Kind:       cfg.Kind,
		AutoDelete: cfg.AutoDelete,
		Durable:    cfg.Durable,
	}
	bnd := cony.Binding{
		Queue:    que,
		Exchange: exc,
		Key:      cfg.Key,
	}
	client.Declare([]cony.Declaration{
		cony.DeclareQueue(que),
		cony.DeclareExchange(exc),
		cony.DeclareBinding(bnd),
	})

	// Declare and register a consumer
	cns := cony.NewConsumer(
		que,
	)
	client.Consume(cns)

	return &subscriber{
		config:       cfg,
		client:       client,
		consumer:     cns,
		stop:         make(chan chan error, 1),
		errors:       make(chan error, 100),
		unmarshaller: unmarshaller,
	}, nil
}

func (s *subscriber) Start() <-chan pubsub.Message {
	output := make(chan pubsub.Message)
	go func(s *subscriber, output chan pubsub.Message) {
		defer func() {
			close(output)
		}()
		for s.client.Loop() {
			select {
			case stop := <-s.stop:
				s.client.Close()
				stop <- nil
				return
			case msg := <-s.consumer.Deliveries():
				output <- &message{
					unmarshaller: s.unmarshaller,
					body:         msg.Body,
					ack:          msg.Ack,
					nack:         msg.Nack,
				}
			case err := <-s.consumer.Errors():
				s.errors <- err
			case err := <-s.client.Errors():
				s.errors <- err
			}
		}
	}(s, output)
	return output
}

func (s *subscriber) Errors() <-chan error {
	return s.errors
}

func (s *subscriber) Stop() error {
	stop := make(chan error)
	s.stop <- stop
	err := <-stop
	return err
}

func (m *message) UnMarshal(msg interface{}) error {
	if m.unmarshaller == nil {
		return pubsub.ErrorNoUnMarshaller
	}
	return m.unmarshaller(m.body, msg)
}

func (m *message) Done() error {
	return m.ack(false)
}
