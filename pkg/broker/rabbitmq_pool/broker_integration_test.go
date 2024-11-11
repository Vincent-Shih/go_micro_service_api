package rabbitmq_test

import (
	"context"
	rabbitmq "go_micro_service_api/pkg/broker/rabbitmq_pool"
	"go_micro_service_api/pkg/cus_err"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRabbitMQ(t *testing.T) {
	// TODO: fill username and password
	user, pass := "", ""
	if user == "" || pass == "" {
		t.Skip("No RabbitMQ credentials provided. Skipping integration tests.")
	}

	broker, kgsErr := rabbitmq.NewBroker(user, pass, "localhost", 5672)
	require.Nil(t, kgsErr)
	defer broker.Close()

	t.Run("TestCreateExchange no set kind", func(t *testing.T) {
		kgsErr := broker.CreateExchange("test", "", false)
		assert.NotNil(t, kgsErr)
		assert.Equal(t, cus_err.InternalServerError, kgsErr.Code().Int())
	})

	t.Run("TestCreateExchange", func(t *testing.T) {
		kgsErr := broker.CreateExchange("test", "direct", false)
		assert.Nil(t, kgsErr)
	})

	t.Run("TestCreateQueue", func(t *testing.T) {
		kgsErr := broker.CreateQueue("test", false)
		assert.Nil(t, kgsErr)
	})

	t.Run("TestBindQueueToExchange no set exchange name", func(t *testing.T) {
		kgsErr := broker.BindQueueToExchange("test", "", "test")
		assert.NotNil(t, kgsErr)
		assert.Equal(t, cus_err.InternalServerError, kgsErr.Code().Int())
	})

	t.Run("TestBindQueueToExchange", func(t *testing.T) {
		kgsErr := broker.BindQueueToExchange("test", "test", "test")
		assert.Nil(t, kgsErr)
	})

	t.Run("TestPublish", func(t *testing.T) {
		ctx := context.Background()
		kgsErr := broker.Publish(ctx, "test", "test", false, []byte("test"))
		assert.Nil(t, kgsErr)
	})

	t.Run("TestConsume", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		delivery, kgsErr := broker.Consume(ctx, "test", "test")
		assert.Nil(t, kgsErr)
		for msg := range delivery {
			broker.Ack(&msg)
		}
	})

	t.Run("DeleteExchange", func(t *testing.T) {
		kgsErr := broker.DeleteExchange("test")
		assert.Nil(t, kgsErr)
	})

	t.Run("DeleteQueue", func(t *testing.T) {
		kgsErr := broker.DeleteQueue("test")
		assert.Nil(t, kgsErr)
	})
}
