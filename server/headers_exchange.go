package server

import (
	"strings"

	"github.com/Michael-kyalo/hare.git/message"
)

// HeadersExchange implements a headers exchange type.
type HeadersExchange struct {
	bindings map[string]queueList
}

type queueList []*Queue

// NewHeadersExchange creates a new HeadersExchange.
func NewHeadersExchange() *HeadersExchange {
	return &HeadersExchange{
		bindings: make(map[string]queueList),
	}
}

// Bind binds a queue to the exchange with header matching rules.
func (ex *HeadersExchange) Bind(q *Queue, routingKey string) {
	headers := message.ExtractHeaders(routingKey)
	xMatch, ok := headers["x-match"]
	if !ok {
		return // "x-match" header is required
	}
	// Create a unique key based on the xMatch and headers
	bindingKey := createBindingKey(xMatch, headers)
	ex.bindings[bindingKey] = append(ex.bindings[bindingKey], q)
}

// Route routes a message to queues based on header matching.
func (ex *HeadersExchange) Route(msg *message.Message, routingKey string) {
	for bindingKey, queues := range ex.bindings {
		// Parse the bindingKey to retrieve xMatch and headers
		xMatch, headers := parseBindingKey(bindingKey)
		if headerMatches(xMatch, headers, msg.Headers) {
			for _, q := range queues {
				q.Enqueue(msg)
			}
		}
	}
}

// createBindingKey creates a unique string key for binding headers.
func createBindingKey(xMatch string, headers map[string]string) string {
	key := xMatch
	for k, v := range headers {
		if k != "x-match" {
			key += "|" + k + "=" + v
		}
	}
	return key
}

// parseBindingKey parses the unique key back into xMatch and headers.
func parseBindingKey(bindingKey string) (string, map[string]string) {
	parts := strings.Split(bindingKey, "|")
	xMatch := parts[0]
	headers := make(map[string]string)

	for _, part := range parts[1:] {
		kv := strings.Split(part, "=")
		if len(kv) == 2 {
			headers[kv[0]] = kv[1]
		}
	}
	return xMatch, headers
}

// headerMatches checks if a message's headers match the binding headers.
func headerMatches(xMatch string, bindingHeaders map[string]string, msgHeaders map[string]string) bool {
	switch xMatch {
	case "all":
		for k, v := range bindingHeaders {
			if msgHeaders[k] != v {
				return false // All headers must match
			}
		}
		return true
	case "any":
		for k, v := range bindingHeaders {
			if msgHeaders[k] == v {
				return true // At least one header must match
			}
		}
		return false
	default:
		return false // Invalid "x-match" value
	}
}
