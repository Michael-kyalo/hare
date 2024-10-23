package connection

import (
	"net"
	"testing"
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
	if c.conn != conn {
		t.Error("Connection's conn field not set correctly")
	}
}
