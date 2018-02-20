package main

import (
	"context"
	"encoding/json"
	"github.com/cheebo/pubsub/memory"
	"log"
	"time"
)

const (
	exchange = "exchange"
	max      = 10
)

func main() {
	hub := memory.NewHub()
	pub := memory.NewPublisher(hub.NewExchange(exchange), json.Marshal)

	sub, err := hub.NewSubscriber(exchange, json.Unmarshal)
	if err != nil {
		log.Fatal(err.Error())
	}

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
