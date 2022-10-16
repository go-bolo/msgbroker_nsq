package msgbroker_nsq

import "github.com/go-catupiry/msgbroker"

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
