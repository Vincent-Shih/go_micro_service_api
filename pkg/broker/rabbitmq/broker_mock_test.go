package rabbitmq_test

import (
	"context"
	"go_micro_service_api/pkg/broker/rabbitmq"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMockRabbitMQ(t *testing.T) {
	t.Parallel()

	mockConn := new(rabbitmq.MockRabbitMQConn)
	mockChannel := new(rabbitmq.MockRabbitMQChannel)
	mockConn.On("Close").Return(nil)
	mockChannel.On("Close").Return(nil)

	broker := rabbitmq.NewBroker(mockConn)
	defer broker.Close()

	t.Run("TestGetConn", func(t *testing.T) {
		conn := broker.GetConn()
		if conn == nil {
			t.Error("Failed to get connection")
		}

		assert.NotNil(t, conn)
	})

	t.Run("TestOpenChannel", func(t *testing.T) {
		mockConn.On("Channel").Return(&amqp.Channel{}, nil)
		_, err := broker.OpenChannel()
		if err != nil {
			t.Errorf("Failed to open channel: %v", err)
		}

		assert.Nil(t, err)
	})

	t.Run("TestCreateExchange", func(t *testing.T) {
		mockChannel.On("ExchangeDeclare", "test", "test", true, false, false, false, mock.Anything).Return(nil)
		err := broker.CreateExchange(mockChannel, "test", "test", true)
		if err != nil {
			t.Errorf("Failed to create exchange: %v", err)
		}
		defer mockChannel.Close()

		assert.Nil(t, err)
	})

	t.Run("TestCreateQueue", func(t *testing.T) {
		mockChannel.On("QueueDeclare", "test", true, false, false, false, mock.Anything).Return(amqp.Queue{}, nil)
		err := broker.CreateQueue(mockChannel, "test", true)
		if err != nil {
			t.Errorf("Failed to create queue: %v", err)
		}
		defer mockChannel.Close()

		assert.Nil(t, err)
	})

	t.Run("TestBindQueueToExchange", func(t *testing.T) {
		mockChannel.On("QueueBind", "test", "test", "test", false, mock.Anything).Return(nil)

		err := broker.BindQueueToExchange(mockChannel, "test", "test", "test")
		if err != nil {
			t.Errorf("Failed to bind queue to exchange: %v", err)
		}
		defer mockChannel.Close()

		assert.Nil(t, err)
	})

	t.Run("TestPublish", func(t *testing.T) {
		mockChannel.On("PublishWithContext", mock.Anything, "test", "test", false, false, mock.Anything).Return(nil)

		ctx := context.Background()
		err := broker.Publish(ctx, mockChannel, "test", "test", false, []byte("test"))
		if err != nil {
			t.Errorf("Failed to publish message: %v", err)
		}
		defer mockChannel.Close()

		assert.Nil(t, err)
	})

	t.Run("TestConsume", func(t *testing.T) {
		mockChannel.On("Qos", 1, 0, true).Return(nil)
		mockChannel.On("ConsumeWithContext", mock.Anything, "test", "", false, false, false, false, mock.Anything).Return(make(<-chan amqp.Delivery), nil)

		ctx := context.Background()
		_, err := broker.Consume(ctx, mockChannel, "test")
		if err != nil {
			t.Errorf("Failed to consume message: %v", err)
		}
		defer mockChannel.Close()

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
