package msgbroker_nsq

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-catupiry/catu"
	"github.com/go-catupiry/msgbroker"
	"github.com/nsqio/go-nsq"
)

func NewNSQClient(cfg *NSQClientCfg) *NSQClient {
	client := NSQClient{
		Logger:          &loggerLogrus{},
		AutoCreateTopic: cfg.AutoCreateTopic,
	}

	if cfg.Config != nil {
		client.Config = cfg.Config
		client.LogLevel = cfg.LogLevel
	}

	return &client
}

type NSQClientCfg struct {
	Config          *nsq.Config
	LogLevel        nsq.LogLevel
	AutoCreateTopic bool
}

type NSQClient struct {
	App             catu.App
	Config          *nsq.Config
	Queues          map[string]msgbroker.Queue
	Producer        *nsq.Producer
	LogLevel        nsq.LogLevel
	Logger          *loggerLogrus
	AutoCreateTopic bool
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
		return err
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
		return "", err
	}

	consumer.SetLogger(c.Logger, c.LogLevel)

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

	if c.AutoCreateTopic {
		// create topic if not exists
		c.CreateTopic(queueName)
	}

	// Use nsqlookupd to discover nsqd instances.
	// See also ConnectToNSQD, ConnectToNSQDs, ConnectToNSQLookupds.
	err = consumer.ConnectToNSQLookupd(addr)
	if err != nil {
		return "", err
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

func (c *NSQClient) ConnectToProducer() error {
	cfgs := c.App.GetConfiguration()
	config := nsq.NewConfig()

	addr := cfgs.GetF("NSQ_ADDR", "127.0.0.1:4150")

	producer, err := nsq.NewProducer(addr, config)
	if err != nil {
		return err
	}

	producer.SetLogger(c.Logger, c.LogLevel)

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

	// Gracefully stop the producer when appropriate (e.g. before shutting down the service)
	// producer.Stop()
	return nil
}

func (c *NSQClient) CreateTopic(name string) error {
	cfgs := c.App.GetConfiguration()
	addr := cfgs.GetF("NSQ_LOOKUPD_ADDR", "127.0.0.1:4161")

	resp, err := http.Post("http://"+addr+"/topic/create?topic="+name, "text/plain", nil)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	return nil
}
