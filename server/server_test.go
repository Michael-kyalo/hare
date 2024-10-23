package server

import (
	"encoding/json"
	"net"
	"testing"
	"time"

	"github.com/Michael-kyalo/hare.git/message"
)

func TestServerCreateExchange(t *testing.T) {
	s := NewServer()
	err := s.CreateExchange("test-exchange", "direct")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := s.exchanges["test-exchange"]; !ok {
		t.Error("Exchange not created")
	}
}

func TestServerCreateQueue(t *testing.T) {
	s := NewServer()
	err := s.CreateQueue("test-queue")
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := s.queues["test-queue"]; !ok {
		t.Error("Queue not created")
	}
}

func TestServerBindQueue(t *testing.T) {
	s := NewServer()
	s.CreateExchange("test-exchange", "direct")
	s.CreateQueue("test-queue")
	err := s.BindQueue("test-queue", "test-exchange", "test-routing-key")
	if err != nil {
		t.Fatal(err)
	}

	// Todo Add assertions to check if the queue is bound to the exchange correctly
	// (This will depend on how we store bindings in your server)
	// Todo add storage logic
}

func TestServerPublish(t *testing.T) {
	s := NewServer()
	s.CreateExchange("test-exchange", "direct")
	s.CreateQueue("test-queue")
	s.BindQueue("test-queue", "test-exchange", "test-routing-key")

	msg := message.NewMessage([]byte("test message"), nil)

	done := make(chan bool)

	go func() {
		err := s.Publish("test-exchange", "test-routing-key", msg)
		if err != nil {
			t.Error(err) // Report the error but don't fail the test immediately
		}
		done <- true
	}()

	go func() {
		q, ok := s.queues["test-queue"]
		if !ok {
			t.Error("Queue not found") // Report the error but don't fail the test immediately
		}
		dequeuedMsg := q.Dequeue()
		if dequeuedMsg == nil || string(dequeuedMsg.Body) != "test message" {
			t.Error("Message not published to the queue")
		}
		done <- true
	}()

	<-done
	<-done // Wait for both goroutines to finish
}

func TestServerAcceptConnection(t *testing.T) {
	s := NewServer()

	// Start the server in a goroutine
	go func() {
		err := s.Start("127.0.0.1:0") // Use an available port
		if err != nil {
			t.Error(err)
		}
	}()

	// Wait for the server to start (you might need a more robust way to do this)
	time.Sleep(10 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Give the server some time to accept the connection
	time.Sleep(10 * time.Millisecond)

	// Assertions (you might need to adjust these based on your server implementation)
	if s.connections == nil || len(s.connections) != 1 {
		t.Error("Server did not accept the connection")
	}
}

func TestServerHandleConnection(t *testing.T) {
	s := NewServer()
	go func() {
		err := s.Start("127.0.0.1:0") // Use an available port
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(10 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send a message (you'll need to define a message format/protocol)
	msg := message.NewMessage([]byte("test message"), nil)
	encoder := json.NewEncoder(conn)
	err = encoder.Encode(msg)
	if err != nil {
		t.Fatal(err)
	}

	// Give the server some time to handle the message
	time.Sleep(10 * time.Millisecond)

	// To do make Assertions
	// For example, check if the message was routed to a queue, etc.
}

func TestServerBind(t *testing.T) {
	s := NewServer()
	s.CreateExchange("test-exchange", "direct")
	s.CreateQueue("test-queue")

	// Start the server
	go func() {
		err := s.Start("127.0.0.1:0")
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(10 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send the BIND command
	bindCmd := []byte("BIND QUEUE test-queue test-exchange test-routing-key")
	_, err = conn.Write(bindCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Assertions (you'll need to add assertions to check if the queue is bound correctly)
	// ...
}

func TestServerPublishConsume(t *testing.T) {
	s := NewServer()
	s.CreateExchange("test-exchange", "direct")
	s.CreateQueue("test-queue")
	s.BindQueue("test-queue", "test-exchange", "test-routing-key")

	// Start the server
	go func() {
		err := s.Start("127.0.0.1:0")
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(10 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Publish a message
	publishCmd := []byte("PUBLISH test-exchange test-routing-key.header1:value1.header2:value2 This is the message body")
	_, err = conn.Write(publishCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Consume the message
	consumeCmd := []byte("CONSUME test-queue")
	_, err = conn.Write(consumeCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Read the consumed message
	var receivedMsg message.Message
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&receivedMsg)
	if err != nil {
		t.Fatal(err)
	}

	// Send ACK
	ackCmd := []byte("ACK " + receivedMsg.ID)
	_, err = conn.Write(ackCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Assertions
	if string(receivedMsg.Body) != "test-message" {
		t.Errorf("Expected 'test-message', got %s", string(receivedMsg.Body))
	}
}

func TestServerConsumeRedelivery(t *testing.T) {
	s := NewServer()
	s.CreateExchange("test-exchange", "direct")
	s.CreateQueue("test-queue")
	s.BindQueue("test-queue", "test-exchange", "test-routing-key")

	// Publish a message
	msg := message.NewMessage([]byte("test message"), nil)
	s.Publish("test-exchange", "test-routing-key", msg)

	// Start the server
	go func() {
		err := s.Start("127.0.0.1:0")
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(10 * time.Millisecond)

	// Connect to the server
	conn, err := net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	// Send the CONSUME command
	consumeCmd := []byte("CONSUME test-queue")
	_, err = conn.Write(consumeCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Read the consumed message (but don't send ACK)
	var receivedMsg message.Message
	decoder := json.NewDecoder(conn)
	err = decoder.Decode(&receivedMsg)
	if err != nil {
		t.Fatal(err)
	}

	// Close the connection to simulate ACK failure
	conn.Close()

	// Reconnect to the server
	conn, err = net.Dial("tcp", s.listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send the CONSUME command again
	consumeCmd = []byte("CONSUME test-queue")
	_, err = conn.Write(consumeCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Read the redelivered message
	err = decoder.Decode(&receivedMsg)
	if err != nil {
		t.Fatal(err)
	}

	// Send ACK
	ackCmd := []byte("ACK " + receivedMsg.ID)
	_, err = conn.Write(ackCmd)
	if err != nil {
		t.Fatal(err)
	}

	// Assertions
	if receivedMsg.DeliveryCount != 2 {
		t.Errorf("Expected DeliveryCount to be 2, got %d", receivedMsg.DeliveryCount)
	}
}
