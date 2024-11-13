package rabbitmq

import (
	"context"
	"fmt"
	"go_micro_service_api/pkg/broker/rabbitmq_pool/internal"
	"go_micro_service_api/pkg/cus_err"
	"net/url"
	"strconv"

	amqp "github.com/rabbitmq/amqp091-go"
)

type brokerImpl struct {
	pool *internal.ChannelPool
}

var _ Broker = (*brokerImpl)(nil)

// NewBroker initializes a new Broker instance with the provided credentials and configuration options.
// This function is designed to be used at server startup and can be directly bound to a dependency injection container.
// It creates a connection and a channel pool, and optionally sets up exchanges, queues, and bindings if a mapping is provided.
//
// Usage:
//
//	broker,err := NewBroker("user", "pass", "localhost", 5672)
//
// Parameters:
//   - user: The username for the RabbitMQ connection.
//   - pass: The password for the RabbitMQ connection.
//   - host: The hostname of the RabbitMQ server.
//   - port: The port number of the RabbitMQ server.
//   - opts: Optional configuration options.
//
// Returns:
//   - *Broker: A pointer to the initialized Broker instance.
//   - *cus_err.CusError: An error if the connection fails.
func NewBroker(user string, pass string, host string, port int, opts ...Option) (Broker, *cus_err.CusError) {
	cfg := &config{}
	for _, opt := range opts {
		opt.apply(cfg)
	}

	dsn := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%d", host, port),
		RawQuery: url.Values{
			"min_connection":     []string{strconv.Itoa(cfg.minConnection)},
			"max_connection":     []string{strconv.Itoa(cfg.maxConnection)},
			"connection_timeout": []string{strconv.Itoa(cfg.connectionTimeout)},
		}.Encode(),
	}

	pool, cusErr := internal.NewChannelPool(dsn.String())
	if cusErr != nil {
		return nil, cusErr
	}

	b := &brokerImpl{
		pool: pool,
	}

	if cfg.mapping == nil {
		return b, nil
	}

	// Create exchanges, queues and binds,when mapping is provided
	for _, opt := range cfg.mapping.Exchanges {
		if err := b.CreateExchange(opt.Name, opt.Kind, opt.Durable); err != nil {
			return nil, err
		}
	}

	for _, opt := range cfg.mapping.Queues {
		if err := b.CreateQueue(opt.Name, opt.Durable); err != nil {
			return nil, err
		}
	}

	for _, opt := range cfg.mapping.Binds {
		if err := b.BindQueueToExchange(opt.QueueName, opt.ExchangeName, opt.RoutingKey); err != nil {
			return nil, err
		}
	}

	return b, nil
}

func (b *brokerImpl) Close() *cus_err.CusError {
	if cusErr := b.pool.Close(); cusErr != nil {
		return cusErr
	}
	return nil
}

func (b *brokerImpl) getChannel() (*amqp.Channel, *cus_err.CusError) {
	ch, cusErr := b.pool.Get()
	if ch == nil {
		return nil, cus_err.New(cus_err.InternalServerError, "Failed to get channel", nil)
	}
	if cusErr != nil {
		return nil, cusErr
	}

	return ch, nil
}

func (b *brokerImpl) CreateExchange(exchange, kind string, durable bool) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	if err := ch.ExchangeDeclare(exchange, kind, durable, false, false, false, nil); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to create exchange", err)
	}

	return nil
}

func (b *brokerImpl) CreateQueue(queue string, durable bool) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	if _, err := ch.QueueDeclare(queue, durable, false, false, false, nil); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to create queue", err)
	}

	return nil
}

func (b *brokerImpl) BindQueueToExchange(queue string, exchange string, routingKey string) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	if err := ch.QueueBind(queue, routingKey, exchange, false, nil); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to bind queue to exchange", err)
	}

	return nil
}

func (b *brokerImpl) Publish(ctx context.Context, exchange string, routingKey string, durable bool, msg []byte) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	mode := amqp.Transient
	if durable {
		mode = amqp.Persistent
	}

	publishing := amqp.Publishing{
		DeliveryMode: mode,
		ContentType:  "application/json",
		Body:         msg,
	}

	if err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, publishing); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to publish message", err)
	}
	return nil
}

func (b *brokerImpl) Consume(ctx context.Context, consumerName string, queueName string) (<-chan amqp.Delivery, *cus_err.CusError) {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return nil, cusErr
	}
	defer b.pool.Put(ch)

	// all of consumers consume one message at a time
	ch.Qos(1, 0, true)

	delivery, err := ch.ConsumeWithContext(
		ctx,
		queueName,
		consumerName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, cus_err.New(cus_err.InternalServerError, "Failed to consume message", err)
	}

	return delivery, nil
}

func (b *brokerImpl) Ack(d *amqp.Delivery) *cus_err.CusError {
	if err := d.Ack(false); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to ack message", err)
	}
	return nil
}

// ExchangeDelete removes the named exchange from the server. When an exchange is
// deleted all queue bindings on the exchange are also deleted.  If this exchange
// does not exist, the channel will be closed with an error.
func (b *brokerImpl) DeleteExchange(exchange string) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	if err := ch.ExchangeDelete(exchange, false, false); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to delete exchange", err)
	}

	return nil
}

// QueueDelete removes the named queue from the server. When a queue is deleted
// all bindings on the queue are also deleted.  If this queue does not exist, the
// channel will be closed with an error.
func (b *brokerImpl) DeleteQueue(queue string) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	if _, err := ch.QueueDelete(queue, false, false, false); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to delete queue", err)
	}

	return nil
}

// UnbindQueueFromExchange removes the binding between an exchange and a queue.
// If the binding does not exist, the channel will be closed with an error.
func (b *brokerImpl) UnbindQueueFromExchange(queue string, exchange string, routingKey string) *cus_err.CusError {
	ch, cusErr := b.getChannel()
	if cusErr != nil {
		return cusErr
	}
	defer b.pool.Put(ch)

	if err := ch.QueueUnbind(queue, routingKey, exchange, nil); err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to unbind queue from exchange", err)
	}

	return nil
}
