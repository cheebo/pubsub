# PubSub

Pub/Sub interface and implementations for
- Dummy: dummy implementation without message transmission
- Mem: exchange in app memory
- RabbitMQ

Todo implementations
- Kafka
- ZMQ

# Interface

Published - publish the messages
Subscriber - listen for messages
Message - interface that wraps messages that receives the Subscriber

Interfaces provide methods to register functions for Marshaling and UnMarshaling.

### Publishing
1. Create new Published and attach Marshaller to it
2. Publish you variable foo with Publisher.Publish(foo). Marshalling happens in Publisher.Publish method

### Subscription
1. Create new Subscriber
2. Attach unmarshaller
3. Listen for incoming messages on channel from Subscriber.Start
4. Each message could be unmarshalled by calling message.UnMarshal(&structure)

```go
type Publisher interface {
	Marshaller(marshaller Marshaller)
	Publish(context.Context, string, interface{}) error
	Errors() <-chan error
}

type Subscriber interface {
	UnMarshaller(unmarshaller UnMarshaller)
	Start() <-chan Message
	Errors() <-chan error
	Stop() error
}

type Message interface {
	UnMarshal(interface{}) error
	Done() error
}
```

# Contributors

1. Sergey Chebotar - design / rabbitmq implementation
2. Stan Shulga - examples for rabbitmq
