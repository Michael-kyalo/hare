package connection

import "net"

// Connection represents a connection to a client.
type Connection struct {
	conn net.Conn
}

// NewConnection creates a new Connection.
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}
