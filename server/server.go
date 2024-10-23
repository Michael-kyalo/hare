package server

import (
	"fmt"

	"github.com/Michael-kyalo/hare.git/message"
)

// Server represents the message queue server.
type Server struct {
	exchanges map[string]Exchange
	queues    map[string]*Queue
}

// NewServer creates a new Server.
func NewServer() *Server {
	return &Server{
		exchanges: make(map[string]Exchange),
		queues:    make(map[string]*Queue),
	}
}

// CreateExchange creates a new exchange with the given name and type.
func (s *Server) CreateExchange(name, typ string) error {
	switch typ {
	case "direct":
		s.exchanges[name] = NewDirectExchange()
	case "fanout":
		s.exchanges[name] = NewFanoutExchange()
	case "topic":
		s.exchanges[name] = NewTopicExchange()
	case "headers":
		s.exchanges[name] = NewHeadersExchange()
	default:
		return fmt.Errorf("invalid exchange type: %s", typ)
	}
	return nil
}

// CreateQueue creates a new queue with the given name.
func (s *Server) CreateQueue(name string) error {
	s.queues[name] = NewQueue(name)
	return nil
}

// BindQueue binds a queue to an exchange with a routing key.
func (s *Server) BindQueue(queueName, exchangeName, routingKey string) error {
	ex, ok := s.exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("exchange not found: %s", exchangeName)
	}
	q, ok := s.queues[queueName]
	if !ok {
		return fmt.Errorf("queue not found: %s", queueName)
	}
	ex.Bind(q, routingKey)
	return nil
}

// Publish publishes a message to the specified exchange with a routing key.
func (s *Server) Publish(exchangeName, routingKey string, msg *message.Message) error {
	ex, ok := s.exchanges[exchangeName]
	if !ok {
		return fmt.Errorf("exchange not found: %s", exchangeName)
	}
	ex.Route(msg, routingKey)
	return nil
}
