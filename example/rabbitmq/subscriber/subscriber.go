package main

import (
	"flag"
	"log"

	"github.com/cheebo/go-config/types"
	"github.com/cheebo/pubsub/example/rabbitmq/message"
	"github.com/cheebo/pubsub/rabbitmq"
	"github.com/golang/protobuf/proto"
)

var url = flag.String("url", "amqp://guest:guest@localhost/", "amqp url")

func main() {
	flag.Parse()

	cfg := new(types.AMQPConfig)
	cfg.URL = *url
	cfg.Queue = "example_queue"
	cfg.Durable = true
	cfg.Exchange = "example_exchange"
	cfg.Kind = "fanout"
	cfg.Key = "example_key"

	sub, err := rabbitmq.NewSubscriber(cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for err := range sub.Errors() {
			log.Println(err)
			sub.Stop()
		}
	}()
	message := new(message.Message)
	for msg := range sub.Start() {
		if err = proto.Unmarshal(msg.Message(), message); err != nil {
			log.Fatal(err)
		}
		if err = msg.Done(); err != nil {
			log.Fatal(err)
		}
		log.Println("Got:", message.ID)
	}
}
