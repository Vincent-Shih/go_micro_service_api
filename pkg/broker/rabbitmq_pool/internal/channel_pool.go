package internal

import (
	"go_micro_service_api/pkg/cus_err"
	"log"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// ChannelPool is a pool of AMQP channels.
// It is used to manage the connection and channels to the RabbitMQ server.
type ChannelPool struct {
	mu          sync.RWMutex     // Mutex for connection status
	pool        *sync.Pool       // Channel pool
	conn        *amqp.Connection // mq connection
	isConnected atomic.Bool      // Connection status flag
	connStr     string           // Connection string
	isReConnect atomic.Bool      // Reconnect flag
	closeChan   chan *amqp.Error // The signal for connection close
	connCond    *sync.Cond       // Condition variable for connection status
}

// NewChannelPool creates a new ChannelPool instance.
func NewChannelPool(connStr string) (*ChannelPool, *cus_err.CusError) {
	cp := &ChannelPool{
		connStr: connStr,
	}

	cp.connCond = sync.NewCond(&cp.mu)

	// Create connection
	if err := cp.connect(); err != nil {
		return nil, err
	}

	// Handle reconnect
	go cp.handleReconnect()

	return cp, nil
}

// connect create a new connection to the RabbitMQ server.
func (p *ChannelPool) connect() *cus_err.CusError {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closeChan = make(chan *amqp.Error)

	conn, err := amqp.Dial(p.connStr)
	if err != nil {
		return cus_err.New(cus_err.InternalServerError, "Failed to connect to RabbitMQ", err)
	}

	// Create channel pool
	p.pool = &sync.Pool{
		New: func() any {
			ch, err := p.conn.Channel()
			if err != nil {
				return nil
			}
			ch.Qos(1, 0, false)
			return ch
		},
	}

	p.conn = conn
	p.isConnected.Store(true)
	p.isReConnect.Store(true)

	// Notify all waiting goroutines that connection is available
	p.connCond.Broadcast()

	// Listen for connection close
	p.conn.NotifyClose(p.closeChan)

	return nil
}

// handleReconnect handles the reconnection to the RabbitMQ server.
func (p *ChannelPool) handleReconnect() {
	for {
		// Wait for connection close
		reason := <-p.closeChan
		if !p.isReConnect.Load() {
			return
		}

		p.isConnected.Store(false)
		log.Printf("Connection closed: %v", reason)

		backoff := time.Second
		maxBackoff := 10 * time.Second

		for p.isReConnect.Load() {
			log.Printf("Attempting to reconnect in %v...", backoff)
			time.Sleep(backoff)

			if err := p.connect(); err != nil {
				log.Printf("Failed to reconnect: %v", err)
				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}

			log.Println("Reconnected successfully")
			p.isConnected.Store(true)
			break
		}
	}
}

// Get retrieves an available AMQP channel from the pool. If the connection is not available,
// this method will block the thread until the connection is re-established. This method is
// thread-safe and ensures that only one thread can access the pool at a time.
func (p *ChannelPool) Get() (*amqp.Channel, *cus_err.CusError) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for !p.isConnected.Load() {
		if !p.isReConnect.Load() {
			return nil, cus_err.New(cus_err.InternalServerError, "Failed to get channel: connection is closed", nil)
		}
		p.connCond.Wait()
	}

	ch, ok := p.pool.Get().(*amqp.Channel)
	if !ok || ch.IsClosed() {
		newCh, err := p.conn.Channel()
		if err != nil {
			return nil, cus_err.New(cus_err.InternalServerError, "Failed to get channel", err)
		}
		newCh.Qos(1, 0, false)
		return newCh, nil
	}

	return ch, nil
}

func (p *ChannelPool) Put(ch *amqp.Channel) {
	if ch == nil || ch.IsClosed() {
		return
	}
	p.pool.Put(ch)
}

func (p *ChannelPool) Close() *cus_err.CusError {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.isReConnect.Store(false)

	// Wake up all waiting goroutines
	p.connCond.Broadcast()

	if p.isConnected.Load() {
		if err := p.conn.Close(); err != nil {
			return cus_err.New(cus_err.InternalServerError, "Failed to close connection", err)
		}
		p.isConnected.Store(false)
	}

	return nil
}
