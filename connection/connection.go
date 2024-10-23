package connection

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Michael-kyalo/hare.git/message"
)

// Connection represents a connection to a client.
type Connection struct {
	Conn net.Conn
}

// HeartbeatMessage is the message used for heartbeats.
type HeartbeatMessage struct {
	Type string `json:"type"` //  field to identify heartbeat messages
}

// NewHeartbeatMessage creates a new HeartbeatMessage.
func NewHeartbeatMessage() *HeartbeatMessage {
	return &HeartbeatMessage{Type: "heartbeat"}
}

// NewConnection creates a new Connection.
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn: conn,
	}
}

// Send sends a message to the client.
func (c *Connection) Send(msg *message.Message) error {
	encoder := json.NewEncoder(c.Conn)
	err := encoder.Encode(msg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err) // Wrap the error
	}
	return nil
}

// Receive receives a message from the client.
func (c *Connection) Receive() (*message.Message, error) {
	reader := bufio.NewReader(c.Conn)
	rawMsg, err := reader.ReadString('\n') // Read until newline character
	if err != nil {
		if err == io.EOF {
			return nil, err // Return EOF for a clean disconnect
		}
		return nil, fmt.Errorf("failed to receive message: %w", err)
	}

	return &message.Message{Body: []byte(rawMsg)}, nil
}

// Close closes the connection.
func (c *Connection) Close() error {
	if c.Conn != nil {
		err := c.Conn.Close()
		c.Conn = nil // Set conn to nil to prevent double-closing
		return err
	}
	return nil
}

// StartHeartbeat starts sending heartbeat messages to the client.
func (c *Connection) StartHeartbeat(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			// You might want to define a specific heartbeat message type
			heartbeatMsg := &message.Message{Body: []byte("heartbeat")}
			if err := c.Send(heartbeatMsg); err != nil {
				fmt.Printf("Failed to send heartbeat message: %v\n", err)
				// For now, we'll just break the loop
				break
			}
		}
	}()
}
