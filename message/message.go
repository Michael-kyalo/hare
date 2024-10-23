package message

import (
	"encoding/json"
	"strings"
)

// Message represents a message with headers and a body
type Message struct {
	ID            string            `json:"id"`
	Headers       map[string]string `json:"headers"`
	Body          []byte            `json:"body"`
	DeliveryCount int               `json:"delivery_count"`
}

// NewMessage creates a new message with the given optional headers and body
func NewMessage(body []byte, headers map[string]string) *Message {
	if headers == nil {
		headers = make(map[string]string)
	}
	return &Message{
		Headers: headers,
		Body:    body,
	}
}

// Marshal marshals the message into the JSON format
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// Unmarshal unmarshals the JSON data into a message
func Unmarshal(data []byte) (*Message, error) {
	var msg Message
	err := json.Unmarshal(data, &msg)
	return &msg, err
}

// ExtractHeaders extracts headers from a routing key
func ExtractHeaders(routingKey string) map[string]string {
	headers := make(map[string]string)
	parts := splitRoutingKey(routingKey)
	for _, part := range parts {
		kv := splitKeyValuePair(part)
		headers[kv[0]] = kv[1]
	}
	return headers
}

// splitRoutingKey splits a routing key into its individual parts
func splitRoutingKey(routingKey string) []string {
	return strings.Split(routingKey, ".")
}

// splitKeyValuePair splits a key-value pair into its individual parts
func splitKeyValuePair(pair string) []string {
	parts := strings.SplitN(pair, ":", 2)
	if len(parts) == 1 {
		return []string{parts[0], ""}
	}
	return parts
}
