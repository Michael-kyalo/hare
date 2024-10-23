package server

import (
	"testing"

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
