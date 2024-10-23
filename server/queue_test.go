package server

import (
	"reflect"
	"testing"

	"github.com/Michael-kyalo/hare.git/message"
)

func TestQueueEnqueueDequeue(t *testing.T) {
	q := NewQueue("testqueue")
	msg := message.NewMessage([]byte("Hello, World!"), nil)

	// Start a goroutine to dequeue the message
	done := make(chan bool) // Channel to signal when done
	go func() {
		dequeuedMsg := q.Dequeue()
		if !reflect.DeepEqual(dequeuedMsg.Body, msg.Body) {
			t.Errorf("Expected dequeued message body to be equal, got %s", string(dequeuedMsg.Body))
		}
		done <- true // Signal that the message was consumed
	}()

	q.Enqueue(msg)
	<-done // Wait for the message to be consumed
}

func TestQueueEnqueueDequeueBlocksWhenEmpty(t *testing.T) {
	q := NewQueue("testqueue")

	select {
	case <-q.messages:
		t.Error("Dequeue should block  when queue is empty")
	default:
		break

	}

}
