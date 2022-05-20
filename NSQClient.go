package msgbroker_nsq

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/msgbroker"
	"github.com/nsqio/go-nsq"
)

func NewNSQClient(cfg *NSQClientCfg) *NSQClient {
	client := NSQClient{}

	if cfg.Config != nil {
		client.Config = cfg.Config
		client.LogLevel = cfg.LogLevel
	}

	return &client
}

type NSQClientCfg struct {
	Config   *nsq.Config
	LogLevel nsq.LogLevel
}

type NSQClient struct {
	App      catu.App
	Config   *nsq.Config
	Queues   map[string]msgbroker.Queue
	Producer *nsq.Producer
	LogLevel nsq.LogLevel
}

func (c *NSQClient) Init(app catu.App) error {
	c.App = app
	cfgs := app.GetConfiguration()

	if c.Config == nil {
		c.Config = nsq.NewConfig()
	}

	addr := cfgs.GetF("NSQ_ADDR", "127.0.0.1:4150")

	producer, err := nsq.NewProducer(addr, c.Config)
	if err != nil {
		log.Fatal(err)
	}

	c.Producer = producer

	return nil
}

// Register a handler function to receive the new messages
// Now we support only one method for each queues
func (c *NSQClient) Subscribe(queueName string, handler msgbroker.MessageHandler) (string, error) {
	cfgs := c.App.GetConfiguration()
	consumer, err := nsq.NewConsumer(queueName, queueName, c.Config)
	if err != nil {
		log.Fatal(err)
	}

	if c.LogLevel != 0 {
		consumer.SetLoggerLevel(c.LogLevel)
	}

	// Set the Handler for messages received by this Consumer. Can be called multiple times.
	// See also AddConcurrentHandlers.
	consumer.AddHandler(&nsqHandlerWrapper{
		QueueName: queueName,
		Handler:   handler,
	})

	addr := cfgs.GetF("NSQ_LOOKUPD_ADDR", "127.0.0.1:4161")
	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(addr)
	if err != nil {
		log.Fatal(err)
	}

	return "", nil
}

type nsqHandlerWrapper struct {
	QueueName string
	Handler   msgbroker.MessageHandler
}

func (h *nsqHandlerWrapper) HandleMessage(message *nsq.Message) error {
	return h.Handler.HandleMessage(h.QueueName, &NSQMessage{Data: &message.Body})
}

// Unsubscribe handler function with subscriberID
func (c *NSQClient) UnSubscribe(subscriberID string) {
	panic("not implemented") // TODO: Implement
}

// Publish one message to queue
func (c *NSQClient) Publish(queueName string, data []byte) error {
	err := c.Producer.Publish(queueName, data)
	if err != nil {
		return err
	}

	return nil
}

// Get queue by queueName
func (c *NSQClient) GetQueue(name string) msgbroker.Queue {
	return c.Queues[name]
}

// Set one queue in queue list
func (c *NSQClient) SetQueue(name string, queue msgbroker.Queue) error {
	c.Queues[name] = queue
	return nil
}

type NSQQueue struct {
	Name    string
	Handler msgbroker.MessageHandler
}

type NSQQueueHandler struct {
}

func (q *NSQQueue) GetName() string {
	return q.Name
}

func (q *NSQQueue) SetName(name string) error {
	q.Name = name
	return nil
}

func (q *NSQQueue) GetHandler() msgbroker.MessageHandler {
	return q.Handler
}

func (q *NSQQueue) SetHandler(handler msgbroker.MessageHandler) error {
	q.Handler = handler
	return nil
}

type NSQMessage struct {
	Data *[]byte
}

func (m *NSQMessage) GetData() *[]byte {
	return m.Data
}

func (c *NSQClient) ConnectToProducer() error {
	config := nsq.NewConfig()
	producer, err := nsq.NewProducer("127.0.0.1:4150", config)
	if err != nil {
		log.Fatal(err)
	}

	c.Producer = producer

	return nil
}

func (c *NSQClient) Close() error {
	log.Println("TODO!")

	// wait for signal to exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Gracefully stop the consumer.
	// consumer.Stop()

	// // Gracefully stop the producer when appropriate (e.g. before shutting down the service)
	// producer.Stop()
	return nil
}
