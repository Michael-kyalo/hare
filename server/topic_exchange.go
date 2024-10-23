package server

import (
	"strings"

	"github.com/Michael-kyalo/hare.git/message"
)

// TopicExchange implements a topic exchange type.
type TopicExchange struct {
	bindings map[string][]*Queue
}

// NewTopicExchange creates a new TopicExchange.
func NewTopicExchange() *TopicExchange {
	return &TopicExchange{
		bindings: make(map[string][]*Queue),
	}
}

// Bind binds a queue to the exchange with a routing key pattern.
func (ex *TopicExchange) Bind(q *Queue, routingKey string) {
	ex.bindings[routingKey] = append(ex.bindings[routingKey], q)
}

// Route routes a message to queues based on matching routing key patterns.
func (ex *TopicExchange) Route(msg *message.Message, routingKey string) {
	for pattern, queues := range ex.bindings {
		if topicMatches(pattern, routingKey) {
			for _, q := range queues {
				q.Enqueue(msg)
			}
		}
	}
}

// topicMatches checks if a routing key matches a topic pattern.
func topicMatches(pattern, routingKey string) bool {
	patternParts := strings.Split(pattern, ".")
	routingKeyParts := strings.Split(routingKey, ".")

	if len(patternParts) != len(routingKeyParts) {
		return false
	}

	for i, part := range patternParts {
		if part == "#" { // Matches any number of words
			return true
		}
		if part != "*" && part != routingKeyParts[i] { // Exact match or "*" (single word)
			return false
		}
	}

	return true
}
