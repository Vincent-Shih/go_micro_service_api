package main

import (
	"context"
	"go_micro_service_api/pkg/broker/rabbitmq"
	"sync"
)

func produce(c context.Context, broker *rabbitmq.Broker, wg *sync.WaitGroup) {
	defer wg.Done()
	ch, err := broker.OpenChannel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	err = broker.Publish(c, ch, "notifications", "", false, []byte("hello"))
	if err != nil {
		panic(err)
	}
}

func main() {
	c := context.Background()
	m := &rabbitmq.BrokerMapping{
		Exchanges: []rabbitmq.ExchangeOpt{
			{
				Name:    "notifications",
				Kind:    "fanout",
				Durable: false,
			},
		},
		Queues: []rabbitmq.QueueOpt{
			{
				Name:    "notification_priority_high",
				Durable: false,
			},
			{
				Name:    "notification_priority_low",
				Durable: false,
			},
		},
		Binds: []rabbitmq.BindOpt{
			{
				QueueName:    "notification_priority_high",
				ExchangeName: "notifications",
				RoutingKey:   "high",
			},
			{
				QueueName:    "notification_priority_low",
				ExchangeName: "notifications",
				RoutingKey:   "low",
			},
		},
	}

	broker, err := rabbitmq.InitBroker(c, "user", "pass", "localhost", 5672, m)
	if err != nil {
		panic(err)
	}
	defer broker.Close()

	var wg sync.WaitGroup

	// publish
	go produce(c, broker, &wg)

	wg.Add(1)
	wg.Wait()
}
