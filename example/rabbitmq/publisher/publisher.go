package main

import (
	"flag"
	"log"
	"time"

	"github.com/cheebo/go-config/types"
	"github.com/cheebo/pubsub/example/rabbitmq/message"
	"github.com/cheebo/pubsub/rabbitmq"
	"encoding/json"
)

var url = flag.String("url", "amqp://guest:guest@localhost/", "amqp url")

func main() {
	flag.Parse()

	cfg := new(types.AMQPConfig)
	cfg.URL = *url
	cfg.Durable = true
	cfg.Exchange = "example_exchange"
	cfg.Kind = "fanout"
	cfg.Key = "example_key"

	pub, err := rabbitmq.NewPublisher(cfg)
	if err != nil {
		log.Fatal(err)
	}
	pub.Marshaller(json.Marshal)

	go func() {
		for err := range pub.Errors() {
			log.Println(err)
		}
	}()

	msg := new(message.Message)
	for i := 0; i <= 10; i++ {
		time.Sleep(time.Second)
		msg.ID = uint64(i)
		err = pub.Publish(nil, "example_key", msg)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Sent:", i)
	}

}
