package main

import (
	"context"
	"encoding/json"
	"github.com/cheebo/pubsub/mem"
	"log"
	"time"
)

const (
	exchange = "exchange"
	max      = 10
)

func main() {
	hub := mem.NewHub()
	pub := mem.NewPublisher(hub.NewExchange(exchange))
	pub.Marshaller(json.Marshal)

	sub, err := hub.NewSubscriber(exchange)
	if err != nil {
		log.Fatal(err.Error())
	}
	sub.UnMarshaller(json.Unmarshal)

	go func() {
		for i := 1; i < max; i++ {
			ctx := context.Background()
			pub.Publish(ctx, "", i)
			time.Sleep(time.Second)
		}
	}()

	for i := range sub.Start() {
		var num int
		err := i.UnMarshal(&num)
		if err != nil {
			log.Fatal(err.Error())
		}
		println("subscriber: ", num)
	}
}
