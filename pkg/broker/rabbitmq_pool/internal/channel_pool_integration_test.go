package internal

import (
	"fmt"
	"go_micro_service_api/pkg/cus_err"
	"os"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Change this to your RabbitMQ URL
// if it is not available, the tests will be skipped
const testAMQPURL = "amqp://admin:admin@localhost:5672/"

func checkRabbitMQAvailable() bool {
	conn, err := amqp.Dial(testAMQPURL)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func TestMain(m *testing.M) {
	if !checkRabbitMQAvailable() {
		fmt.Println("RabbitMQ is not available, skipping tests")
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNewChannelPool(t *testing.T) {
	tests := []struct {
		name            string
		connStr         string
		wantErr         bool
		exceptedCusCode cus_err.CusCode
	}{
		{
			name:    "Valid connection string",
			connStr: testAMQPURL,
			wantErr: false,
		},
		{
			name:            "Invalid connection string",
			connStr:         "invalid://localhost",
			wantErr:         true,
			exceptedCusCode: cus_err.InternalServerError,
		},
		{
			name:            "Empty connection string",
			connStr:         "",
			wantErr:         true,
			exceptedCusCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewChannelPool(tt.connStr)
			if tt.wantErr {
				assert.Nil(t, pool)
				assert.NotNil(t, err)
				assert.Equal(t, tt.exceptedCusCode, err.Code())
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, pool)
				assert.True(t, pool.isConnected.Load())
				// Close pool
				defer pool.Close()
			}
		})
	}
}

func TestChannelPool_Get(t *testing.T) {
	pool, err := NewChannelPool(testAMQPURL)
	require.Nil(t, err)
	defer pool.Close()

	t.Run("Get channel successfully", func(t *testing.T) {
		ch, err := pool.Get()
		assert.Nil(t, err)
		assert.NotNil(t, ch)
		assert.False(t, ch.IsClosed())
		pool.Put(ch)
	})

	t.Run("Get multiple channels", func(t *testing.T) {
		ch1, err1 := pool.Get()
		ch2, err2 := pool.Get()

		assert.Nil(t, err1)
		assert.Nil(t, err2)
		assert.NotNil(t, ch1)
		assert.NotNil(t, ch2)
		assert.NotEqual(t, ch1, ch2)

		pool.Put(ch1)
		pool.Put(ch2)
	})

	t.Run("Get channel after connection close", func(t *testing.T) {
		ch1, err1 := pool.Get()
		assert.Nil(t, err1)
		assert.NotNil(t, ch1)
		pool.Put(ch1)

		// Close the pool
		cusErr := pool.Close()
		require.Nil(t, cusErr)

		// Get channel after close should not return any channel
		ch, err := pool.Get()
		assert.NotNil(t, err)
		assert.Nil(t, ch)
	})
}

func TestChannelPool_Put(t *testing.T) {
	pool, err := NewChannelPool(testAMQPURL)
	assert.Nil(t, err)
	defer pool.Close()

	t.Run("Put valid channel", func(t *testing.T) {
		ch, err := pool.Get()
		assert.Nil(t, err)

		pool.Put(ch)
		// Check the the channel can be retrieved from the pool
		ch2 := pool.pool.Get().(*amqp.Channel)
		assert.Equal(t, ch, ch2)
	})

	t.Run("Put nil channel", func(t *testing.T) {
		// Should not panic
		pool.Put(nil)
	})

	t.Run("Put closed channel", func(t *testing.T) {
		ch, err := pool.Get()
		assert.Nil(t, err)

		// Should put the channel back to the pool
		ch.Close()
		pool.Put(ch)

		// Verify that the channel is not in the pool
		ch2 := pool.pool.Get().(*amqp.Channel)
		assert.NotEqual(t, ch, ch2)
	})
}

func TestChannelPool_Close(t *testing.T) {
	pool, err := NewChannelPool(testAMQPURL)
	assert.Nil(t, err)

	t.Run("Close pool", func(t *testing.T) {
		err := pool.Close()
		assert.Nil(t, err)
		assert.True(t, pool.conn.IsClosed())
		assert.False(t, pool.isConnected.Load())
		assert.False(t, pool.isReConnect.Load())
	})

	t.Run("Double close", func(t *testing.T) {
		err := pool.Close()
		assert.Nil(t, err)
		err = pool.Close()
		assert.Nil(t, err)
	})
}

func TestChannelPool_Reconnect(t *testing.T) {
	pool, err := NewChannelPool(testAMQPURL)
	assert.Nil(t, err)
	defer pool.Close()

	t.Run("Test reconnect", func(t *testing.T) {
		// Close connection to force reconnect
		pool.conn.Close()

		// Get a channel
		ch, err := pool.Get()
		require.Nil(t, err)
		require.NotNil(t, ch)
	})
}

func TestChannelPool_Concurrent(t *testing.T) {
	pool, err := NewChannelPool(testAMQPURL)
	assert.Nil(t, err)
	defer pool.Close()

	t.Run("Concurrent channel operations", func(t *testing.T) {
		workers := 10
		iterations := 100
		done := make(chan bool)

		for i := 0; i < workers; i++ {
			go func() {
				for j := 0; j < iterations; j++ {
					ch, err := pool.Get()
					assert.Nil(t, err)
					assert.NotNil(t, ch)

					time.Sleep(time.Millisecond)
					pool.Put(ch)
				}
				done <- true
			}()
		}

		// Wait for all workers to finish
		for i := 0; i < workers; i++ {
			<-done
		}
	})
}

func TestChannelPool_ErrorCases(t *testing.T) {
	t.Run("Invalid connection string", func(t *testing.T) {
		pool, err := NewChannelPool("amqp://invalid:5672/")
		assert.NotNil(t, err)
		assert.Nil(t, pool)
	})

	pool, err := NewChannelPool(testAMQPURL)
	assert.Nil(t, err)
	defer pool.Close()

	t.Run("Get channel after close", func(t *testing.T) {
		pool.Close()
		ch, err := pool.Get()
		assert.Error(t, err)
		assert.Nil(t, ch)
	})
}
