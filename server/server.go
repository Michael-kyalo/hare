package server

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/Michael-kyalo/hare.git/connection"
	"github.com/Michael-kyalo/hare.git/message"
	"github.com/google/uuid"
)

// Server represents the message queue server.
type Server struct {
	exchanges   map[string]Exchange
	queues      map[string]*Queue
	listener    net.Listener
	connections map[net.Conn]*connection.Connection // Store active connections
}

// NewServer creates a new Server.
func NewServer() *Server {
	return &Server{
		exchanges:   make(map[string]Exchange),
		queues:      make(map[string]*Queue),
		connections: make(map[net.Conn]*connection.Connection),
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

// Start starts the server.
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.listener = listener
	fmt.Println("Server listening on", listener.Addr())

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		c := connection.NewConnection(conn)
		s.connections[conn] = c

		go s.handleConnection(c) // Handle the connection in a separate goroutine
	}
}

// handleConnection handles communication with a client.

func (s *Server) handleConnection(c *connection.Connection) {
	defer func() {
		delete(s.connections, c.Conn)
		c.Close()
	}()

	for {
		msg, err := c.Receive()
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error receiving message:", err)
			}
			break
		}

		// Process the message
		err = s.processMessage(c, msg)
		if err != nil {
			fmt.Println("Error processing message:", err)
		}
	}
}
func (s *Server) processMessage(c *connection.Connection, msg *message.Message) error {
	parts := strings.Split(string(msg.Body), " ")
	command := parts[0]

	switch command {
	case "PUBLISH":
		return s.handlePublish(parts)
	case "CONSUME":
		return s.handleConsume(c, parts)
	case "ACK":
		return s.handleAck(parts)
	case "CREATE":
		return s.handleCreate(parts)
	case "BIND":
		return s.handleBind(parts)
	default:
		return fmt.Errorf("invalid command: %s", command)
	}
}

// handlePublish handles publish command
func (s *Server) handlePublish(parts []string) error {
	if len(parts) != 4 {
		return fmt.Errorf("invalid PUBLISH command format")
	}
	exchange := parts[1]
	routingKey := parts[2]
	msgBody := []byte(parts[3])

	return s.Publish(exchange, routingKey, message.NewMessage(msgBody, nil))
}

// handleConsume handles the consume command
func (s *Server) handleConsume(c *connection.Connection, parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("invalid CONSUME command format")
	}
	queueName := parts[1]

	q, ok := s.queues[queueName]
	if !ok {
		return fmt.Errorf("queue not found: %s", queueName)
	}

	for {
		msg := q.Dequeue()
		if msg == nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		msg.ID = generateMessageID()
		msg.DeliveryCount++

		if err := c.Send(msg); err != nil {
			return fmt.Errorf("error sending message to client: %w", err)
		}

		ackMsg, err := c.Receive()
		if err != nil {
			q.Enqueue(msg) // Re-enqueue on error
			return fmt.Errorf("error receiving ACK: %w", err)
		}

		if string(ackMsg.Body) != "ACK "+msg.ID { // Check ACK with message ID
			q.Enqueue(msg) // Re-enqueue on non-ACK or incorrect ID
			return fmt.Errorf("unexpected response from client: %s", string(ackMsg.Body))
		}

		// Message acknowledged
	}
}

// handleBind handles the BIND command
func (s *Server) handleBind(parts []string) error {
	if len(parts) != 5 || parts[1] != "QUEUE" {
		return fmt.Errorf("invalid BIND command format")
	}
	queueName := parts[2]
	exchangeName := parts[3]
	routingKey := parts[4]

	return s.BindQueue(queueName, exchangeName, routingKey)
}

// handleAck handles ACK command
func (s *Server) handleAck(parts []string) error {
	if len(parts) != 2 {
		return fmt.Errorf("invalid ACK command format")
	}
	messageID := parts[1]

	// TODO:: add logic here to track acknowledged messages.
	fmt.Println("Acknowledged message:", messageID)
	return nil
}

// handleCreate handles the create command
func (s *Server) handleCreate(parts []string) error {
	if len(parts) < 3 {
		return fmt.Errorf("invalid CREATE command format")
	}
	subcommand := parts[1]
	name := parts[2]

	switch subcommand {
	case "EXCHANGE":
		if len(parts) != 4 {
			return fmt.Errorf("invalid CREATE EXCHANGE command format")
		}
		typ := parts[3]
		return s.CreateExchange(name, typ)
	case "QUEUE":
		return s.CreateQueue(name)
	default:
		return fmt.Errorf("invalid CREATE subcommand: %s", subcommand)
	}
}

// generateMessageID generates a unique message ID using UUIDs.
func generateMessageID() string {
	return uuid.New().String()
}
