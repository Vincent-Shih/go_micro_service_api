# RabbitMQ Broker

作為 rabbitmq 的抽象層

## Usage

在應用程式中使用時，可採下列步驟初始化

首先定義該應用程式要使用的 exchange 跟 queue

!!! note exchange & queue
    exchange 的功能類似 router，根據規則發送訊息到對應的 queue

例如，以下範例定義一個名為 `notification` 的 exchange，兩個 queue 分別名為 `notification_priority_high` 跟 `notification_priority_low` ， 並將兩者綁定，由於該 exchange 是 `direct` 模式會完全匹配key來判斷要傳到哪個綁定的queue

```go
m := &rabbitmq.BrokerMapping{
    Exchanges: []rabbitmq.ExchangeOpt{
        {
            Name:    "notification",
            Kind:    "direct",
            Durable: false,
        },
    },
    Queues: []rabbitmq.QueueOpt{
        {
            Name:    "notification_priority_high",
            Durable: false,
        },
        {
            Name:    "notification_priority_low",
            Durable: false,
        },
    },
    Binds: []rabbitmq.BindOpt{
        {
            QueueName:    "notification_priority_high",
            ExchangeName: "notification",
            RoutingKey:   "high",
        },
        {
            QueueName:    "notification_priority_low",
            ExchangeName: "notification",
            RoutingKey:   "low",
        },
    },
}
```

初始化 broker，傳入連線訊息與剛剛的路由設定

!!! danger 用完要記得關閉連線

```go
c := context.Background()
broker, err := rabbitmq.InitBroker(c, "user", "pass", "localhost", 5672, m)
if err != nil {
    panic(err)
}
defer broker.Close()
```

### 發送訊息

先開通道

!!! danger 通道使用完需關閉

以下範例會發送訊息到名為`notification`的 exchange，該 exchange 服從 `direct` 模式 會轉發訊息到有著相同pattern "high" 的 queue

```go
// publish
go func() {
    ch, err := broker.OpenChannel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()

    err = broker.Puslish(c, ch, "notification", "high", false, []byte("hello"))
    if err != nil {
        panic(err)
    }
}()
```

### 讀取訊息

先開通道

!!! danger 通道使用完需關閉

以下範例會讀取訊息從名為`notification_priority_high`的 queue

!!! warning 收到訊息後要呼叫 `Ack` ，不然訊息會一直堆在 rabbitmq

```go
// consume
go func() {
    ch, err := broker.OpenChannel()
    if err != nil {
        panic(err)
    }
    defer ch.Close()

    d, err := broker.Consume(c, ch, "notification_priority_high")
    if err != nil {
        panic(err)
    }

    for msg := range d {
        println(string(msg.Body))
		broker.Ack(&msg)
    }
}()
```

以上範例程式碼皆在 `examples/direct`

```sh

go run examples/direct/producer/main.go

go run examples/direct/consumer/main.go
hello
```