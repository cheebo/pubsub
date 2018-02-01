package main

import (
	"flag"
	"log"

	"encoding/json"
	"github.com/cheebo/go-config/types"
	"github.com/cheebo/pubsub/example/rabbitmq/message"
	"github.com/cheebo/pubsub/rabbitmq"
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
	sub.UnMarshaller(json.Unmarshal)

	go func() {
		for err := range sub.Errors() {
			log.Println(err)
			sub.Stop()
		}
	}()
	for msg := range sub.Start() {
		message := new(message.Message)
		msg.UnMarshal(message)
		if err = msg.UnMarshal(message); err != nil {
			log.Fatal(err)
		}
		if err = msg.Done(); err != nil {
			log.Fatal(err)
		}
		log.Println("Got:", message.ID)
	}
}
