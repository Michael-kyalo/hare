package connection

import (
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Michael-kyalo/hare.git/message"
)

func TestNewConnection(t *testing.T) {
	// Create a mock listener (we won't actually listen for connections)
	listener, err := net.Listen("tcp", "127.0.0.1:0") // Use an available port
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Simulate a client connection (without actually connecting)
	conn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Create a Connection
	c := NewConnection(conn)

	// Assertions
	if c.Conn != conn {
		t.Error("Connection's conn field not set correctly")
	}
}

func TestConnectionSendReceive(t *testing.T) {
	// Create a mock connection using pipes
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	// Create a Connection
	c := NewConnection(serverConn)

	// Send a message
	msg := message.NewMessage([]byte("test message"), nil)
	err := c.Send(msg)
	if err != nil {
		t.Fatal(err)
	}

	// Receive the message (on the client side)
	var receivedMsg message.Message
	decoder := json.NewDecoder(clientConn)
	err = decoder.Decode(&receivedMsg)
	if err != nil {
		t.Fatal(err)
	}

	// Assertions
	if string(receivedMsg.Body) != "test message" {
		t.Errorf("Expected 'test message', got %s", string(receivedMsg.Body))
	}
}
func TestConnectionSendReceiveError(t *testing.T) {
	// Create a mock connection using pipes
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	// Create a Connection
	c := NewConnection(serverConn)

	// Send a message
	msg := message.NewMessage([]byte("test message"), nil)

	done := make(chan bool)

	go func() {
		err := c.Send(msg)
		if err != nil {
			t.Error(err) // Report error but don't fail immediately
			done <- true
			return
		}

		// Close the client connection to simulate an error
		clientConn.Close()

		// Try to receive a message (should return an error)
		_, err = c.Receive()
		if err == nil {
			t.Error("Expected an error, got nil")
		} else if err != io.EOF { // Check for specific error type (EOF in this case)
			t.Errorf("Expected io.EOF, got %v", err)
		}
		done <- true
	}()

	<-done // Wait for the goroutine to finish
}

func TestConnectionClose(t *testing.T) {
	// Create a mock connection using pipes
	serverConn, clientConn := net.Pipe()

	// Create a Connection
	c := NewConnection(serverConn)

	// Close the connection
	err := c.Close()
	if err != nil {
		t.Errorf("Close() returned an error: %v", err)
	}

	// Try to write to the connection (should return an error)
	_, err = clientConn.Write([]byte("test"))
	if err == nil {
		t.Error("Expected an error after closing the connection, got nil")
	}
}

func TestConnectionHeartbeat(t *testing.T) {
	// Create a mock connection using pipes
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()

	// Create a Connection
	c := NewConnection(serverConn)

	// Start the heartbeat with a short interval
	c.StartHeartbeat(10 * time.Millisecond)

	// Receive a few heartbeat messages
	var receivedMsg message.Message
	decoder := json.NewDecoder(clientConn)
	for i := 0; i < 3; i++ {
		err := decoder.Decode(&receivedMsg)
		if err != nil {
			t.Fatal(err)
		}
		if string(receivedMsg.Body) != "heartbeat" {
			t.Errorf("Expected 'heartbeat', got %s", string(receivedMsg.Body))
		}
	}
}
