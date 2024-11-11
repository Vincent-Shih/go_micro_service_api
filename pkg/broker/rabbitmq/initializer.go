package rabbitmq

import (
	"context"
	"go_micro_service_api/pkg/cus_err"
	"go_micro_service_api/pkg/cus_otel"
)

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

func InitBroker(
	ctx context.Context,
	user string,
	pass string,
	host string,
	port int,
	mapping *BrokerMapping,
	opts ...Option,
) (*Broker, *cus_err.CusError) {
	conn := newConnection(user, pass, host, port, opts...)
	broker := NewBroker(conn)

	ch, err := broker.OpenChannel()
	if err != nil {
		kgsErr := cus_err.New(cus_err.InternalServerError, "Failed to create channel", err)
		cus_otel.Error(ctx, kgsErr.Error())
		return nil, kgsErr
	}
	defer ch.Close()

	if mapping == nil {
		return broker, nil
	}

	// Create exchanges, queues and binds,when mapping is provided
	for _, opt := range mapping.Exchanges {
		if err := broker.CreateExchange(ch, opt.Name, opt.Kind, opt.Durable); err != nil {
			kgsErr := cus_err.New(cus_err.InternalServerError, "Failed to create exchange", err)
			cus_otel.Error(ctx, kgsErr.Error())
			return nil, kgsErr
		}
	}

	for _, opt := range mapping.Queues {
		if err := broker.CreateQueue(ch, opt.Name, opt.Durable); err != nil {
			kgsErr := cus_err.New(cus_err.InternalServerError, "Failed to create queue", err)
			cus_otel.Error(ctx, kgsErr.Error())
			return nil, kgsErr
		}
	}

	for _, opt := range mapping.Binds {
		if err := broker.BindQueueToExchange(ch, opt.QueueName, opt.ExchangeName, opt.RoutingKey); err != nil {
			kgsErr := cus_err.New(cus_err.InternalServerError, "Failed to bind queue to exchange", err)
			cus_otel.Error(ctx, kgsErr.Error())
			return nil, kgsErr
		}
	}

	return broker, nil
}
