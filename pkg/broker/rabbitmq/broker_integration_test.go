package rabbitmq_test

import (
	"context"
	"go_micro_service_api/pkg/broker/rabbitmq"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRabbitMQ(t *testing.T) {
	// TODO: fill username and password
	user, pass := "", ""
	if user == "" || pass == "" {
		t.Skip("No RabbitMQ credentials provided. Skipping integration tests.")
	}

	ctx := context.Background()
	broker, err := rabbitmq.InitBroker(ctx, user, pass, "localhost", 5672, &rabbitmq.BrokerMapping{})
	if err != nil {
		t.Skip("No RabbitMQ connection available. Skipping integration tests.")
	}
	defer broker.Close()

	setUp := func() rabbitmq.RabbitMQChannel {
		ch, _ := broker.OpenChannel()
		broker.CreateExchange(ch, "test", "direct", false)
		broker.CreateQueue(ch, "test", false)
		broker.BindQueueToExchange(ch, "test", "test", "test")

		return ch
	}

	tearDown := func(ch rabbitmq.RabbitMQChannel) {
		ch.ExchangeDelete("test", false, false)
		ch.QueueDelete("test", false, false, false)
		defer ch.Close()
	}

	t.Run("TestGetConn", func(t *testing.T) {
		conn := broker.GetConn()
		if conn == nil {
			t.Error("Failed to get connection")
		}

		assert.NotNil(t, conn)
	})

	t.Run("TestOpenChannel", func(t *testing.T) {
		ch, err := broker.OpenChannel()
		if err != nil {
			t.Errorf("Failed to open channel: %v", err)
		}
		defer ch.Close()

		assert.Nil(t, err)
	})

	t.Run("TestCreateExchange", func(t *testing.T) {
		ch, _ := broker.OpenChannel()
		err := broker.CreateExchange(ch, "test", "direct", false)

		if err != nil {
			t.Errorf("Failed to create exchange: %v", err)
		}
		defer ch.Close()
		defer ch.ExchangeDelete("test", false, false)

		assert.Nil(t, err)
	})

	t.Run("TestCreateQueue", func(t *testing.T) {
		ch, _ := broker.OpenChannel()
		err := broker.CreateQueue(ch, "test", false)

		if err != nil {
			t.Errorf("Failed to create queue: %v", err)
		}
		defer ch.Close()
		defer ch.QueueDelete("test", false, false, false)

		assert.Nil(t, err)
	})

	t.Run("TestBindQueueToExchange", func(t *testing.T) {
		ch, _ := broker.OpenChannel()
		broker.CreateExchange(ch, "test", "direct", false)
		broker.CreateQueue(ch, "test", false)

		err := broker.BindQueueToExchange(ch, "test", "test", "test")
		if err != nil {
			t.Errorf("Failed to bind queue to exchange: %v", err)
		}

		defer tearDown(ch)

		assert.Nil(t, err)
	})

	t.Run("TestPublish", func(t *testing.T) {
		ctx := context.Background()
		ch := setUp()

		err := broker.Publish(ctx, ch, "test", "test", false, []byte("test"))
		if err != nil {
			t.Errorf("Failed to publish message: %v", err)
		}

		defer tearDown(ch)

		assert.Nil(t, err)
	})

	t.Run("TestConsume", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		ch := setUp()

		d, err := broker.Consume(ctx, ch, "test")
		if err != nil {
			t.Errorf("Failed to consume message: %v", err)
		}

		go func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					broker.Publish(ctx, ch, "test", "test", false, []byte("hello"))
					time.Sleep(1 * time.Millisecond)
				}
			}
		}(ctx)

		for msg := range d {
			assert.Equal(t, msg.Body, []byte("hello"))
			broker.Ack(&msg)
			break
		}

		cancel()

		defer tearDown(ch)

		assert.Nil(t, err)
	})

	t.Run("TestClose", func(t *testing.T) {
		err := broker.Close()
		if err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}

		assert.Nil(t, err)
	})
}

func TestCreateMqFailed(t *testing.T) {
	// TODO: fill username and password
	user, pass := "", ""
	if user == "" || pass == "" {
		t.Skip("No RabbitMQ credentials provided. Skipping integration tests.")
	}
	t.Run("Empty Exchange", func(t *testing.T) {
		ctx := context.Background()
		broker, err := rabbitmq.InitBroker(ctx, user, pass, "localhost", 5672, &rabbitmq.BrokerMapping{
			Exchanges: []rabbitmq.ExchangeOpt{
				{
					Name: "",
					Kind: "",
				},
			},
		})
		assert.Nil(t, broker)
		assert.NotNil(t, err)
	})

	t.Run("Empty Bind", func(t *testing.T) {
		ctx := context.Background()
		broker, err := rabbitmq.InitBroker(ctx, user, pass, "localhost", 5672, &rabbitmq.BrokerMapping{
			Binds: []rabbitmq.BindOpt{
				{
					QueueName:    "",
					RoutingKey:   "",
					ExchangeName: "",
				},
			},
		})
		assert.Nil(t, broker)
		assert.NotNil(t, err)
	})
}
