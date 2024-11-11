package rabbitmq

import (
	"context"
	"go_micro_service_api/pkg/cus_err"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Broker interface {
	// Close the connection
	Close() *cus_err.CusError
	// Create an exchange
	CreateExchange(exchange string, kind string, durable bool) *cus_err.CusError
	// Create a queue
	CreateQueue(queue string, durable bool) *cus_err.CusError
	// Bind a queue to an exchange
	BindQueueToExchange(queue string, exchange string, routingKey string) *cus_err.CusError
	// Publish message to an exchange
	Publish(ctx context.Context, exchange string, routingKey string, durable bool, msg []byte) *cus_err.CusError
	// Consume message from a queue
	Consume(ctx context.Context, consumerName string, queueName string) (<-chan amqp.Delivery, *cus_err.CusError)
	// Ack
	Ack(msg *amqp.Delivery) *cus_err.CusError
	// Delete Exchange
	DeleteExchange(exchange string) *cus_err.CusError
	// Delete Queue
	DeleteQueue(queue string) *cus_err.CusError
	// Unbind Queue from Exchange
	UnbindQueueFromExchange(queue string, exchange string, routingKey string) *cus_err.CusError
}
