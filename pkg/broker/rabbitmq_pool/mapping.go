package rabbitmq

type ExchangeOpt struct {
	Name    string
	Kind    string
	Durable bool
}

type QueueOpt struct {
	Name    string
	Durable bool
}

type BindOpt struct {
	QueueName    string
	ExchangeName string
	RoutingKey   string
}

type BrokerMapping struct {
	Exchanges []ExchangeOpt
	Queues    []QueueOpt
	Binds     []BindOpt
}
