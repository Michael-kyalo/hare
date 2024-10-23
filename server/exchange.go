package server

import (
	"github.com/Michael-kyalo/hare.git/message"
)

type Exchange interface {
	Bind(q *Queue, routingKey string)
	Route(msg *message.Message, routingKey string)
}
