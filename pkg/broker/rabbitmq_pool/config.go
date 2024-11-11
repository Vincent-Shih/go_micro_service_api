package rabbitmq

type config struct {
	maxConnection     int
	minConnection     int
	connectionTimeout int
	mapping           *BrokerMapping
}

type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithMinConnection sets the minimum number of connections to the RabbitMQ server.
func WithMinConnection(v int) Option {
	return optionFunc(func(c *config) {
		c.minConnection = v
	})
}

// WithMaxConnection sets the maximum number of connections to the RabbitMQ server.
func WithMaxConnection(v int) Option {
	return optionFunc(func(c *config) {
		c.maxConnection = v
	})
}

// WithConnectionTimeout sets the connection timeout in seconds.
func WithConnectionTimeout(v int) Option {
	return optionFunc(func(c *config) {
		c.connectionTimeout = v
	})
}

// WithMapping sets the mapping for exchanges, queues, and bindings.
//
// Example:
//
//	mapping := &BrokerMapping{
//	  Exchanges: []rabbitmq.ExchangeOpt{
//	    {Name: "exchange1", Kind: "direct", Durable: false}
//	  },
//	  Queues: []rabbitmq.QueueOpt{
//	    {Name: "queue1", Durable: false}
//	  },
//	  Binds: []rabbitmq.BindOpt{
//	    {QueueName: "queue1", ExchangeName: "exchange1", RoutingKey: "key1"}
//	  },
//	}
//	broker,err := NewBroker("user", "pass", "localhost", 5672, WithMapping(mapping))
func WithMapping(v *BrokerMapping) Option {
	return optionFunc(func(c *config) {
		c.mapping = v
	})
}
