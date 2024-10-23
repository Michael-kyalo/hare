package server

import "github.com/Michael-kyalo/hare.git/message"

// DirectExchange implements a direct exchange type.
type DirectExchange struct {
	bindings map[string][]*Queue
}

// NewDirectExchange creates a new DirectExchange.
func NewDirectExchange() *DirectExchange {
	return &DirectExchange{
		bindings: make(map[string][]*Queue),
	}
}

// Bind binds a queue to the exchange with a routing key.
func (ex *DirectExchange) Bind(q *Queue, routingKey string) {
	ex.bindings[routingKey] = append(ex.bindings[routingKey], q)
}

// Route routes a message to the queues based on the routing key.
func (ex *DirectExchange) Route(msg *message.Message, routingKey string) {
	for _, q := range ex.bindings[routingKey] {
		q.Enqueue(msg)
	}
}
