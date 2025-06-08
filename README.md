# NSQ Message Broker Module for Go Bolo Framework

A NSQ message broker implementation for the [Go Bolo Framework](https://github.com/go-bolo/bolo), providing easy-to-use pub/sub messaging capabilities.

## Features

- **Publisher/Subscriber pattern**: Publish and subscribe to NSQ topics
- **Automatic topic creation**: Optional automatic topic creation
- **Message handling**: Flexible message handler interface
- **Multiple publish modes**: Single, multiple, and deferred publishing
- **Connection management**: Automatic connection handling and cleanup
- **Logging integration**: Built-in logging with configurable levels
- **Testing support**: Comprehensive test coverage with mock handlers

## Installation

```bash
go get github.com/go-bolo/msgbroker_nsq
```

## Configuration

Set the following environment variables:

```bash
# NSQ daemon address (default: 127.0.0.1:4150)
NSQ_ADDR=127.0.0.1:4150

# NSQ lookup daemon address (default: 127.0.0.1:4161)
NSQ_LOOKUPD_ADDR=127.0.0.1:4161
```

## Basic Usage

### 1. Initialize the NSQ Client

```go
package main

import (
    "github.com/go-bolo/bolo"
    "github.com/go-bolo/msgbroker"
    "github.com/go-bolo/msgbroker_nsq"
    "github.com/nsqio/go-nsq"
)

func main() {
    app := bolo.Init()

    // Basic configuration
    nsqClient := msgbroker_nsq.NewNSQClient(&msgbroker_nsq.NSQClientCfg{
        AutoCreateTopic: true,
        LogLevel:        nsq.LogLevelInfo,
    })

    // Register the message broker plugin
    app.RegisterPlugin(msgbroker.NewPlugin(&msgbroker.PluginCfgs{
        Client: nsqClient,
    }))

    err := app.Bootstrap()
    if err != nil {
        panic(err)
    }
}
```

### 2. Create a Message Handler

```go
type MyMessageHandler struct{}

func (h *MyMessageHandler) HandleMessage(queueName string, message msgbroker.Message) error {
    data := message.GetData()
    if data != nil {
        fmt.Printf("Received message on queue '%s': %s\n", queueName, string(*data))
    }
    return nil
}
```

### 3. Subscribe to Messages

```go
handler := &MyMessageHandler{}
subscriberID, err := nsqClient.Subscribe("my-topic", handler)
if err != nil {
    log.Fatalf("Failed to subscribe: %v", err)
}

// Later, to unsubscribe
nsqClient.UnSubscribe(subscriberID)
```

### 4. Publishing Messages

```go
// Single message
err := nsqClient.Publish("my-topic", []byte("Hello, World!"))
if err != nil {
    log.Printf("Failed to publish: %v", err)
}

// Multiple messages
messages := [][]byte{
    []byte("Message 1"),
    []byte("Message 2"),
    []byte("Message 3"),
}
err = nsqClient.MultiPublish("my-topic", messages)
if err != nil {
    log.Printf("Failed to multi-publish: %v", err)
}

// Deferred message (delayed delivery)
import "time"
delay := 30 * time.Second
err = nsqClient.DeferredPublish("my-topic", delay, []byte("Delayed message"))
if err != nil {
    log.Printf("Failed to publish deferred message: %v", err)
}
```

## Advanced Configuration

### Custom NSQ Configuration

```go
nsqConfig := nsq.NewConfig()
nsqConfig.MaxInFlight = 1000
nsqConfig.MaxAttempts = 3

nsqClient := msgbroker_nsq.NewNSQClient(&msgbroker_nsq.NSQClientCfg{
    Config:          nsqConfig,
    LogLevel:        nsq.LogLevelDebug,
    AutoCreateTopic: true,
})
```

### Queue Management

```go
// Create and manage queues
queue := &msgbroker_nsq.NSQQueue{
    Name:    "my-queue",
    Handler: &MyMessageHandler{},
}

err := nsqClient.SetQueue("my-queue", queue)
if err != nil {
    log.Printf("Failed to set queue: %v", err)
}

// Retrieve queue
retrievedQueue := nsqClient.GetQueue("my-queue")
```

## Complete Example

```go
package main

import (
    "fmt"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/go-bolo/bolo"
    "github.com/go-bolo/msgbroker"
    "github.com/go-bolo/msgbroker_nsq"
    "github.com/nsqio/go-nsq"
)

type OrderHandler struct{}

func (h *OrderHandler) HandleMessage(queueName string, message msgbroker.Message) error {
    data := message.GetData()
    if data != nil {
        fmt.Printf("Processing order from queue '%s': %s\n", queueName, string(*data))
        // Process your order logic here
    }
    return nil
}

func main() {
    app := bolo.Init()

    // Initialize NSQ client
    nsqClient := msgbroker_nsq.NewNSQClient(&msgbroker_nsq.NSQClientCfg{
        AutoCreateTopic: true,
        LogLevel:        nsq.LogLevelInfo,
    })

    // Register the message broker plugin
    app.RegisterPlugin(msgbroker.NewPlugin(&msgbroker.PluginCfgs{
        Client: nsqClient,
    }))

    err := app.Bootstrap()
    if err != nil {
        log.Fatalf("Failed to bootstrap app: %v", err)
    }

    // Subscribe to orders topic
    orderHandler := &OrderHandler{}
    subscriberID, err := nsqClient.Subscribe("orders", orderHandler)
    if err != nil {
        log.Fatalf("Failed to subscribe to orders: %v", err)
    }

    fmt.Printf("Subscribed to orders topic with ID: %s\n", subscriberID)

    // Publish some test orders
    orders := [][]byte{
        []byte(`{"id": 1, "product": "Widget A", "quantity": 5}`),
        []byte(`{"id": 2, "product": "Widget B", "quantity": 3}`),
        []byte(`{"id": 3, "product": "Widget C", "quantity": 2}`),
    }

    err = nsqClient.MultiPublish("orders", orders)
    if err != nil {
        log.Printf("Failed to publish orders: %v", err)
    }

    // Wait for interrupt signal to gracefully shutdown
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    <-c

    // Cleanup
    nsqClient.UnSubscribe(subscriberID)
    nsqClient.Close()
    fmt.Println("Shutting down...")
}
```

## Testing

The module includes comprehensive tests for all components:

```bash
# Run all tests
go test -v

# Run specific test files
go test -v ./NSQMessage_test.go ./NSQMessage.go
go test -v ./NSQQueue_test.go ./NSQQueue.go

# Run tests with coverage
go test -cover
```

### Creating Mock Handlers for Testing

```go
type MockHandler struct {
    Messages []msgbroker.Message
}

func (m *MockHandler) HandleMessage(queueName string, message msgbroker.Message) error {
    m.Messages = append(m.Messages, message)
    return nil
}

func TestMyFunction(t *testing.T) {
    handler := &MockHandler{}
    
    // Use handler in your tests
    msg := &msgbroker_nsq.NSQMessage{Data: &[]byte("test")}
    err := handler.HandleMessage("test-queue", msg)
    
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    if len(handler.Messages) != 1 {
        t.Errorf("Expected 1 message, got %d", len(handler.Messages))
    }
}
```

## API Reference

### NSQClient Methods

- `Init(app bolo.App) error` - Initialize the client with the app
- `Subscribe(queueName string, handler msgbroker.MessageHandler) (string, error)` - Subscribe to a topic
- `UnSubscribe(subscriberID string)` - Unsubscribe from a topic
- `Publish(queueName string, data []byte) error` - Publish a single message
- `MultiPublish(queueName string, dataList [][]byte) error` - Publish multiple messages
- `DeferredPublish(queueName string, delay time.Duration, data []byte) error` - Publish with delay
- `GetQueue(name string) msgbroker.Queue` - Get a queue by name
- `SetQueue(name string, queue msgbroker.Queue) error` - Set a queue
- `Close() error` - Close all connections and cleanup

### NSQMessage Methods

- `GetData() *[]byte` - Get the message data

### NSQQueue Methods

- `GetName() string` - Get the queue name
- `SetName(name string) error` - Set the queue name
- `GetHandler() msgbroker.MessageHandler` - Get the message handler
- `SetHandler(handler msgbroker.MessageHandler) error` - Set the message handler

## Error Handling

The client provides detailed error messages for common scenarios:

```go
subscriberID, err := nsqClient.Subscribe("my-topic", handler)
if err != nil {
    log.Printf("Subscription failed: %v", err)
    // Handle subscription error
}

err = nsqClient.Publish("my-topic", []byte("message"))
if err != nil {
    log.Printf("Publishing failed: %v", err)
    // Handle publishing error
}
```

## Best Practices

1. **Always handle errors**: Check for errors on all operations
2. **Graceful shutdown**: Always call `Close()` to cleanup connections
3. **Topic naming**: Use descriptive topic names (e.g., "user.created", "order.processed")
4. **Message format**: Consider using JSON for structured data
5. **Error handling in handlers**: Handle errors gracefully in message handlers
6. **Testing**: Write tests for your message handlers using mock objects

## Requirements

- Go 1.18+
- NSQ daemon running
- Go Bolo Framework

## Related Links

- [Go Bolo Framework](https://github.com/go-bolo/bolo)
- [Message Broker Interface](https://github.com/go-bolo/msgbroker)  
- [NSQ Official Documentation](https://nsq.io/)
- [go-nsq Client Library](https://github.com/nsqio/go-nsq)

## License

MIT