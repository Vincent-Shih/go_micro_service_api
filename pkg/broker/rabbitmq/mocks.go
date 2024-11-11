package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type MockRabbitMQConn struct {
	mock.Mock
}

func (m *MockRabbitMQConn) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockRabbitMQConn) Channel() (*amqp.Channel, error) {
	args := m.Called()
	return args.Get(0).(*amqp.Channel), args.Error(1)
}

type MockRabbitMQChannel struct {
	mock.Mock
}

func (m *MockRabbitMQChannel) Qos(prefetchCount, prefetchSize int, global bool) error {
	args := m.Called(prefetchCount, prefetchSize, global)
	return args.Error(0)
}

func (m *MockRabbitMQChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	margs := m.Called(name, kind, durable, autoDelete, internal, noWait, args)
	return margs.Error(0)
}

func (m *MockRabbitMQChannel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error) {
	margs := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return margs.Get(0).(amqp.Queue), margs.Error(1)
}

func (m *MockRabbitMQChannel) QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error) {
	args := m.Called(name, ifUnused, ifEmpty, noWait)
	return args.Int(0), args.Error(1)
}

func (m *MockRabbitMQChannel) ExchangeDelete(name string, ifUnused, noWait bool) error {
	args := m.Called(name, ifUnused, noWait)
	return args.Error(0)
}

func (m *MockRabbitMQChannel) QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error {
	margs := m.Called(name, key, exchange, noWait, args)
	return margs.Error(0)
}

func (m *MockRabbitMQChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func (m *MockRabbitMQChannel) PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	args := m.Called(ctx, exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func (m *MockRabbitMQChannel) ConsumeWithContext(ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	margs := m.Called(ctx, queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return margs.Get(0).(<-chan amqp.Delivery), margs.Error(1)
}

func (m *MockRabbitMQChannel) Close() error {
	args := m.Called()
	return args.Error(0)
}
