package server

import "github.com/Michael-kyalo/hare.git/message"

// FanoutExchange implements a fanout exchange type.
type FanoutExchange struct {
	queues []*Queue
}

// NewFanoutExchange creates a new FanoutExchange.
func NewFanoutExchange() *FanoutExchange {
	return &FanoutExchange{}
}

// Bind binds a queue to the exchange (routing key is ignored).
func (ex *FanoutExchange) Bind(q *Queue, routingKey string) {
	ex.queues = append(ex.queues, q)
}

// Route routes a message to all bound queues.
func (ex *FanoutExchange) Route(msg *message.Message, routingKey string) {
	for _, q := range ex.queues {
		q.Enqueue(msg)
	}
}
