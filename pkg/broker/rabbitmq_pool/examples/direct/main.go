package main

import (
	"context"
	"fmt"
	rabbitmq "go_micro_service_api/pkg/broker/rabbitmq_pool"
	"log"
	"sync"
	"time"
)

// Change this to your RabbitMQ credentials
const (
	user = "admin"
	pass = "admin"
	host = "localhost"
	port = 5672
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Define the mapping
	m := &rabbitmq.BrokerMapping{
		Exchanges: []rabbitmq.ExchangeOpt{
			{
				Name:    "notifications",
				Kind:    "direct", // Set the exchange kind to direct
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

	// Create a new broker
	broker, cusErr := rabbitmq.NewBroker(user, pass, host, port, rabbitmq.WithMapping(m))
	if cusErr != nil {
		log.Fatalf("Failed to create broker: %v", cusErr)
	}
	defer broker.Close()

	// Producer
	go produce(ctx, broker)

	// Consumer
	var wg sync.WaitGroup
	wg.Add(2)
	go high_consume(ctx, broker, &wg)
	go low_consume(ctx, broker, &wg)

	// Wait for 3 seconds to consume the message
	time.Sleep(3 * time.Second)
	cancel()

	wg.Wait()

	// Delete the exchange and queues
	err := broker.DeleteExchange("notifications")
	if err != nil {
		log.Fatalf("Failed to delete exchange: %v", err)
	}

	err = broker.DeleteQueue("notification_priority_high")
	if err != nil {
		log.Fatalf("Failed to delete queue: %v", err)
	}
	err = broker.DeleteQueue("notification_priority_low")
	if err != nil {
		log.Fatalf("Failed to delete queue: %v", err)
	}
}

func produce(ctx context.Context, broker rabbitmq.Broker) {
	// Publish 5 message
	for i := 1; i < 6; i++ {
		// Set the routing key to high if i is even, otherwise set it to low
		routeKey := "high"
		if i%2 == 0 {
			routeKey = "low"
		}

		msg := fmt.Sprintf("hello: %d ", i)
		log.Printf("Publishing message: %s", msg)
		// Publish message to the exchange
		// In fanout exchange, the routing key is ignored so it can be empty
		err := broker.Publish(ctx, "notifications", routeKey, false, []byte(msg))
		if err != nil {
			log.Fatalf("Failed to publish message: %v", err)
		}
	}
}

func high_consume(ctx context.Context, broker rabbitmq.Broker, wg *sync.WaitGroup) {
	defer wg.Done()

	d, cusErr := broker.Consume(ctx, "Consumer1", "notification_priority_high")
	if cusErr != nil {
		log.Fatalf("Failed to consume message: %v", cusErr)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer1: Context done")
			return
		case msg, ok := <-d:
			if !ok {
				log.Println("Consumer1: Channel closed")
				return
			}
			log.Printf("[Consumer1]: Received message: %s", string(msg.Body))
			if err := broker.Ack(&msg); err != nil {
				log.Printf("Consumer1: Failed to ack message: %v", err)
			}
		}
	}
}

func low_consume(ctx context.Context, broker rabbitmq.Broker, wg *sync.WaitGroup) {
	defer wg.Done()

	d, cusErr := broker.Consume(ctx, "Consumer2", "notification_priority_low")
	if cusErr != nil {
		log.Fatalf("Failed to consume message: %v", cusErr)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer2: Context done")
			return
		case msg, ok := <-d:
			if !ok {
				log.Println("Consumer2: Channel closed")
				return
			}
			log.Printf("[Consumer2]: Received message: %s", string(msg.Body))
			if err := broker.Ack(&msg); err != nil {
				log.Printf("Consumer2: Failed to ack message: %v", err)
			}
		}
	}
}
