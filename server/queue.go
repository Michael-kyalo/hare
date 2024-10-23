package server

import "github.com/Michael-kyalo/hare.git/message"

// Queue represents a message queue with an in-memory store
type Queue struct {
	name     string
	messages chan *message.Message
}

// NewQueue creates a new message queue with the given name
func NewQueue(name string) *Queue {
	return &Queue{
		name:     name,
		messages: make(chan *message.Message),
	}
}

// Name returns the name of the queue
func (q *Queue) Name() string {
	return q.name
}

// Enqueue adds a message to the queue
func (q *Queue) Enqueue(msg *message.Message) {
	q.messages <- msg
}

// Dequeue removes and returns the next message from the queue
func (q *Queue) Dequeue() *message.Message {
	return <-q.messages
}
