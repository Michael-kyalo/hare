package server

import (
	"testing"

	"github.com/Michael-kyalo/hare.git/message"
)

func TestExchangeBind(t *testing.T) {
	t.Run("DirectExchange", func(t *testing.T) {
		ex := NewDirectExchange()
		q := NewQueue("test-queue")
		ex.Bind(q, "test-routing-key")

		// Assertion specific to DirectExchange
		if _, ok := ex.bindings["test-routing-key"]; !ok {
			t.Error("Queue not bound to the exchange with the correct routing key")
		}
	})

	t.Run("FanoutExchange", func(t *testing.T) {
		ex := NewFanoutExchange()
		q := NewQueue("test-queue")
		ex.Bind(q, "") // Routing key is ignored for Fanout exchange

		// Assertion specific to FanoutExchange
		if len(ex.queues) != 1 || ex.queues[0] != q {
			t.Error("Queue not bound to the Fanout exchange correctly")
		}
	})

	t.Run("TopicExchange", func(t *testing.T) {
		ex := NewTopicExchange()
		q := NewQueue("test-queue")
		ex.Bind(q, "topic.*")

		// Assertion specific to TopicExchange
		if _, ok := ex.bindings["topic.*"]; !ok {
			t.Error("Queue not bound to the Topic exchange with the correct routing key pattern")
		}
	})

	t.Run("HeadersExchange", func(t *testing.T) {
		ex := NewHeadersExchange()
		q := NewQueue("test-queue")
		headers := map[string]string{
			"x-match": "all",
			"header1": "value1",
			"header2": "value2",
		}
		ex.Bind(q, "")

		// Assertion specific to HeadersExchange
		bindingKey := createBindingKey(headers["x-match"], headers)
		if _, ok := ex.bindings[bindingKey]; !ok {
			t.Error("Queue not bound to the Headers exchange with the correct headers")
		}
	})
}

func TestExchangeRoute(t *testing.T) {
	t.Run("DirectExchange", func(t *testing.T) {
		ex := NewDirectExchange()
		msg := message.NewMessage([]byte("test message"), nil)
		q := NewQueue("test-queue")

		// Start a goroutine to dequeue the message
		done := make(chan bool) //channel to signal when done
		go func() {
			dequeuedMsg := q.Dequeue()
			//Assertions specific to DirectExchange
			if dequeuedMsg == nil || string(dequeuedMsg.Body) != "test message" {
				t.Errorf("Message not routed correctly")
			}
			done <- true // Signal that the message was consumed
		}()

		ex.Bind(q, "test-routing-key")
		ex.Route(msg, "test-routing-key")
		<-done
	})

	t.Run("FanoutExchange", func(t *testing.T) {
		ex := NewFanoutExchange()
		q1 := NewQueue("q1")
		q2 := NewQueue("q2")
		msg := message.NewMessage([]byte("test message"), nil)

		// Start a goroutine to dequeue the message
		done := make(chan bool) // Channel to signal when done
		go func() {
			dequeuedMsg1 := q1.Dequeue()
			//Assertions specific to fanout exchange
			if dequeuedMsg1 == nil || string(dequeuedMsg1.Body) != "test message" {
				t.Error("Message not routed to q1 correctly")
			}
			done <- true // Signal that q1 received the message
		}()

		go func() {
			dequeuedMsg2 := q2.Dequeue()
			//Assertions specific to fanout exchange
			if dequeuedMsg2 == nil || string(dequeuedMsg2.Body) != "test message" {
				t.Error("Message not routed to q2 correctly")
			}
			done <- true // Signal that q2 received the message
		}()

		ex.Bind(q1, "") // Routing key is ignored for Fanout exchange
		ex.Bind(q2, "")
		ex.Route(msg, "") // Routing key is ignored for Fanout exchange

		<-done
		<-done // Wait for both messages to be consumed
	})

	t.Run("TopicExchange", func(t *testing.T) {
		ex := NewTopicExchange()
		q1 := NewQueue("q1")
		q2 := NewQueue("q2")
		msg1 := message.NewMessage([]byte("message 1"), nil)
		msg2 := message.NewMessage([]byte("message 2"), nil)

		// Start a goroutine to dequeue the message
		done := make(chan bool)
		go func() {
			dequeuedMsg1 := q1.Dequeue()
			if dequeuedMsg1 == nil || string(dequeuedMsg1.Body) != "message 1" {
				t.Error("Message 1 not routed to q1 correctly")
			}
			done <- true
		}()

		go func() {
			dequeuedMsg2 := q2.Dequeue()
			if dequeuedMsg2 == nil || string(dequeuedMsg2.Body) != "message 2" {
				t.Error("Message 2 not routed to q2 correctly")
			}
			done <- true
		}()

		ex.Bind(q1, "topic.one.*")
		ex.Bind(q2, "*.two.#")
		ex.Route(msg1, "topic.one.any") // Should go to q1
		ex.Route(msg2, "any.two.other") // Should go to q2
		<-done
		<-done // Wait for both messages to be consumed
	})

	t.Run("HeadersExchange", func(t *testing.T) {
		ex := NewHeadersExchange()
		q1 := NewQueue("q1")
		q2 := NewQueue("q2")

		msg1 := message.NewMessage([]byte("message 1"), map[string]string{"header1": "value1", "header2": "value2"})
		msg2 := message.NewMessage([]byte("message 2"), map[string]string{"header1": "value1", "header3": "value3"})

		// Start a goroutine to dequeue the message
		done := make(chan bool)

		go func() {
			dequeuedMsg1 := q1.Dequeue()
			// Assertions specific to HeadersExchange
			if dequeuedMsg1 == nil || string(dequeuedMsg1.Body) != "message 1" {
				t.Error("Message 1 not routed to q1 correctly")
			}
			done <- true // notify queue
		}()

		go func() {
			dequeuedMsg2 := q2.Dequeue()
			if dequeuedMsg2 == nil || string(dequeuedMsg2.Body) != "message 2" {
				t.Error("Message 2 not routed to q2 correctly")
			}
			done <- true // notify queue
		}()

		ex.Bind(q1, "")
		ex.Bind(q2, "")

		ex.Route(msg1, "") // Should go to q1 (matches all headers)
		ex.Route(msg2, "") // Should go to q2 (matches at least one header)

		<-done
		<-done // Wait for both messages to be consumed
	})
}
