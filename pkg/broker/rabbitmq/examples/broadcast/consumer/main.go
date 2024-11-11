package main

import (
	"context"
	"go_micro_service_api/pkg/broker/rabbitmq"
	"sync"
	"time"
)

func consume(c context.Context, broker *rabbitmq.Broker, queue string, wg *sync.WaitGroup) {
	defer wg.Done()
	ch, err := broker.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	ctx, cancel := context.WithTimeout(c, 5*time.Second)
	defer cancel()

	d, err := broker.Consume(ctx, ch, queue)
	if err != nil {
		panic(err)
	}

	for msg := range d {
		println(string(msg.Body))
		broker.Ack(&msg)
	}
}

func main() {
	var wg sync.WaitGroup

	c := context.Background()
	broker, err := rabbitmq.InitBroker(c, "user", "pass", "localhost", 5672, nil)
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	go consume(c, broker, "notification_priority_high", &wg)
	go consume(c, broker, "notification_priority_low", &wg)

	wg.Add(2)
	wg.Wait()
}
