package message

import (
	"reflect"
	"testing"
)

func TestNewMessage(t *testing.T) {
	headers := map[string]string{"key": "value"}
	body := []byte("Hello, World!")
	msg := NewMessage(body, headers)

	if !reflect.DeepEqual(msg.Headers, headers) {
		t.Errorf("Expected headers to be equal  %v, got %v", headers, msg.Headers)
	}
	if !reflect.DeepEqual(msg.Body, body) {
		t.Errorf("Expected body to be equal %s, got %s", body, string(msg.Body))
	}

}

func TestNewMessageNoHeaders(t *testing.T) {
	body := []byte("Hello, World!")
	msg := NewMessage(body, nil)

	if len(msg.Headers) != 0 {
		t.Errorf("Expected headers to be empty when no headers were provided")
	}
	if !reflect.DeepEqual(msg.Body, body) {
		t.Errorf("Expected body to be equal %s, got %s", body, string(msg.Body))
	}
}
