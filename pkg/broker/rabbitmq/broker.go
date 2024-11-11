package rabbitmq

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type config struct {
	maxConnection     int
	minConnection     int
	connectionTimeout int
}

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

func WithMinConnection(v int) Option {
	return optionFunc(func(c *config) {
		c.minConnection = v
	})
}

func WithMaxConnection(v int) Option {
	return optionFunc(func(c *config) {
		c.maxConnection = v
	})
}

func WithConnectionTimeout(v int) Option {
	return optionFunc(func(c *config) {
		c.connectionTimeout = v
	})
}

type IBroker interface {
	// get the connection
	GetConn() RabbitMQConn
	// close the connection
	Close() error
	// open a channel
	OpenChannel() (RabbitMQChannel, error)
	// create an exchange
	CreateExchange(ch RabbitMQChannel, exchange string, kind string, durable bool) error
	// create a queue
	CreateQueue(ch RabbitMQChannel, queue string, durable bool) error
	// bind a queue to an exchange
	BindQueueToExchange(ch RabbitMQChannel, queue string, exchange string, routingKey string) error
	// publish message to an exchange
	Publish(ctx context.Context, ch RabbitMQChannel, exchange string, routingKey string, durable bool, msg []byte) error
	// consume message from a queue
	Consume(ctx context.Context, ch RabbitMQChannel, queueName string) (<-chan amqp.Delivery, error)
}

type RabbitMQConn interface {
	Channel() (*amqp.Channel, error)
	Close() error
}

type RabbitMQChannel interface {
	Qos(prefetchCount, prefetchSize int, global bool) error
	QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args amqp.Table) (amqp.Queue, error)
	ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error
	QueueDelete(name string, ifUnused, ifEmpty, noWait bool) (int, error)
	ExchangeDelete(name string, ifUnused, noWait bool) error
	QueueBind(name, key, exchange string, noWait bool, args amqp.Table) error
	PublishWithContext(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error
	ConsumeWithContext(ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
	Close() error
}

type Broker struct {
	conn RabbitMQConn
}

var _ IBroker = (*Broker)(nil)

var (
	instance *amqp.Connection
	once     sync.Once
)

func newConnection(user string, pass string, host string, port int, opts ...Option) *amqp.Connection {
	once.Do(func() {
		cfg := config{}
		for _, opt := range opts {
			opt.apply(&cfg)
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

		conn, err := amqp.Dial(dsn.String())
		if err != nil {
			log.Fatalf("Failed to connect to rabbitmq: %v", err)
		}
		instance = conn
	})

	return instance
}

func NewBroker(conn RabbitMQConn) *Broker {
	return &Broker{conn: conn}
}

func (b *Broker) GetConn() RabbitMQConn {
	return b.conn
}

func (b *Broker) Close() error {
	return b.conn.Close()
}

func (b *Broker) OpenChannel() (RabbitMQChannel, error) {
	return b.conn.Channel()
}

func (b *Broker) CreateExchange(ch RabbitMQChannel, exchange string, kind string, durable bool) error {
	return ch.ExchangeDeclare(
		exchange,
		kind,
		durable,
		false,
		false,
		false,
		nil,
	)
}

func (b *Broker) CreateQueue(ch RabbitMQChannel, queue string, durable bool) error {
	_, err := ch.QueueDeclare(
		queue,
		durable,
		false,
		false,
		false,
		nil,
	)
	return err
}

func (b *Broker) BindQueueToExchange(ch RabbitMQChannel, queue string, exchange string, routingKey string) error {
	return ch.QueueBind(
		queue,
		routingKey,
		exchange,
		false,
		nil,
	)
}

func (b *Broker) Publish(ctx context.Context, ch RabbitMQChannel, exchange string, routingKey string, durable bool, msg []byte) error {
	mode := amqp.Transient
	if durable {
		mode = amqp.Persistent
	}

	return ch.PublishWithContext(
		ctx,
		exchange,
		routingKey,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: mode,
			ContentType:  "application/json",
			Body:         msg,
		})
}

func (b *Broker) Consume(ctx context.Context, ch RabbitMQChannel, queueName string) (<-chan amqp.Delivery, error) {
	// all of consumers consume one message at a time
	ch.Qos(1, 0, true)

	return ch.ConsumeWithContext(
		ctx,
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func (b *Broker) Ack(d *amqp.Delivery) error {
	return d.Ack(false)
}
