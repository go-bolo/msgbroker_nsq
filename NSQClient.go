package msgbroker_nsq

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-bolo/bolo"
	"github.com/go-bolo/msgbroker"
	"github.com/nsqio/go-nsq"
)

func NewNSQClient(cfg *NSQClientCfg) msgbroker.Client {
	client := NSQClient{
		Logger:          &loggerLogrus{},
		AutoCreateTopic: cfg.AutoCreateTopic,
		Consumers:       make(map[string]*nsq.Consumer),
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
	App             bolo.App
	Config          *nsq.Config
	Queues          map[string]msgbroker.Queue
	Consumers       map[string]*nsq.Consumer
	Producer        *nsq.Producer
	LogLevel        nsq.LogLevel
	Logger          *loggerLogrus
	AutoCreateTopic bool
}

func (c *NSQClient) Init(app bolo.App) error {
	c.App = app
	cfgs := app.GetConfiguration()

	if c.Config == nil {
		c.Config = nsq.NewConfig()
	}

	if c.Consumers == nil {
		c.Consumers = make(map[string]*nsq.Consumer)
	}

	addr := cfgs.GetF("NSQ_ADDR", "127.0.0.1:4150")

	producer, err := nsq.NewProducer(addr, c.Config)
	if err != nil {
		return fmt.Errorf("Init: failed to create producer: %w", err)
	}

	c.Producer = producer

	return nil
}

func (c *NSQClient) Subscribe(queueName string, handler msgbroker.MessageHandler) (string, error) {
	cfgs := c.App.GetConfiguration()
	consumer, err := nsq.NewConsumer(queueName, queueName, c.Config)
	if err != nil {
		return "", fmt.Errorf("Init: failed to create producer: %w", err)
	}

	consumer.SetLogger(c.Logger, c.LogLevel)

	if c.LogLevel != 0 {
		consumer.SetLoggerLevel(c.LogLevel)
	}

	consumer.AddHandler(&nsqHandlerWrapper{
		QueueName: queueName,
		Handler:   handler,
	})

	addr := cfgs.GetF("NSQ_LOOKUPD_ADDR", "127.0.0.1:4161")

	if c.AutoCreateTopic {
		c.CreateTopic(queueName)
	}

	err = consumer.ConnectToNSQLookupd(addr)
	if err != nil {
		return "", fmt.Errorf("Subscribe: failed to create consumer: %w", err)
	}

	// Store the consumer for tracking and later cleanup
	c.Consumers[queueName] = consumer

	return queueName, nil
}

type nsqHandlerWrapper struct {
	QueueName string
	Handler   msgbroker.MessageHandler
}

func (h *nsqHandlerWrapper) HandleMessage(message *nsq.Message) error {
	return h.Handler.HandleMessage(h.QueueName, &NSQMessage{Data: &message.Body})
}

func (c *NSQClient) UnSubscribe(subscriberID string) {
	if consumer, exists := c.Consumers[subscriberID]; exists {
		consumer.Stop()
		delete(c.Consumers, subscriberID)
	}
}

func (c *NSQClient) Publish(queueName string, data []byte) error {
	err := c.Producer.Publish(queueName, data)
	if err != nil {
		return fmt.Errorf("Publish: failed to publish message: %w", err)
	}

	return nil
}

func (c *NSQClient) MultiPublish(queueName string, dataList [][]byte) error {
	err := c.Producer.MultiPublish(queueName, dataList)
	if err != nil {
		return fmt.Errorf("MultiPublish: failed to publish messages: %w", err)
	}

	return nil
}

func (c *NSQClient) DeferredPublish(queueName string, delay time.Duration, data []byte) error {
	err := c.Producer.DeferredPublish(queueName, delay, data)
	if err != nil {
		return fmt.Errorf("DeferredPublish: failed to publish message: %w", err)
	}

	return nil
}

func (c *NSQClient) GetQueue(name string) msgbroker.Queue {
	return c.Queues[name]
}

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
		return fmt.Errorf("ConnectToProducer: failed to create producer: %w", err)
	}

	producer.SetLogger(c.Logger, c.LogLevel)

	c.Producer = producer

	return nil
}

func (c *NSQClient) Close() error {
	// Stop all consumers
	for subscriberID, consumer := range c.Consumers {
		consumer.Stop()
		delete(c.Consumers, subscriberID)
	}

	// Stop the producer
	if c.Producer != nil {
		c.Producer.Stop()
	}

	return nil
}

func (c *NSQClient) CreateTopic(name string) error {
	cfgs := c.App.GetConfiguration()
	addr := cfgs.GetF("NSQ_LOOKUPD_ADDR", "127.0.0.1:4161")

	resp, err := http.Post("http://"+addr+"/topic/create?topic="+name, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("CreateTopic: failed to create topic '%s': %w", name, err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("CreateTopic: failed to create topic '%s': HTTP status %d %s",
			name, resp.StatusCode, resp.Status)
	}

	return nil
}
