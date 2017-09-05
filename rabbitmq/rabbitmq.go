package rabbitmq

import (
	"context"
	"github.com/cheebo/pubsub"
	"github.com/assembla/cony"
	"github.com/golang/protobuf/proto"
	"github.com/streadway/amqp"
	"time"
	"github.com/cheebo/go-config/types"
)

type publisher struct {
	config *types.AMQPConfig
	client *cony.Client
	publisher *cony.Publisher
}

type subscriber struct {
	config *types.AMQPConfig
	client *cony.Client
	consumer *cony.Consumer
	stop chan chan error
	errors chan error
}

type message struct {
	body []byte
	ack  func(bool) error
	nack func(bool, bool) error

}

func NewPublisher(cfg *types.AMQPConfig) (pubsub.Publisher, error) {
	client := cony.NewClient(
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
	client.Declare([]cony.Declaration{
		cony.DeclareExchange(exc),
	})

	// Declare and register a publisher
	pub := cony.NewPublisher(exc.Name, cfg.Key)
	client.Publish(pub)

	return &publisher{
		config: cfg,
		client: client,
		publisher: pub,
	}, nil
}


func (p *publisher) Publish(ctx context.Context, key string, msg proto.Message) error {
	body, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	publishing := amqp.Publishing{
		DeliveryMode: p.config.DeliveryMode,
		ContentType: "application/x-protobuf",
		Timestamp: time.Now(),
		Body: body,
	}

	if len(key) > 0 {
		return p.publisher.PublishWithRoutingKey(publishing, key)
	} else {
		return p.publisher.Publish(publishing)
	}
}



func NewSubscriber(cfg *types.AMQPConfig) (pubsub.Subscriber, error) {
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
		config: cfg,
		client: client,
		consumer: cns,
		stop: make(chan chan error, 1),
		errors: make(chan error, 100),
	}, nil
}

func (s *subscriber) Start() <-chan pubsub.Message {
	output := make(chan pubsub.Message)
	go func(s *subscriber, output chan pubsub.Message) {
		defer close(output)
		for s.client.Loop() {
			select {
			case stop := <-s.stop:
				s.client.Close()
				stop <- nil
				return
			case msg := <-s.consumer.Deliveries():
				output <- &message{
					body: msg.Body,
					ack: msg.Ack,
					nack: msg.Nack,
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
	err := <- stop
	return err
}


func (m *message) Message() []byte {
	return m.body
}

func (m *message) Done() error {
	return m.ack(false)
}