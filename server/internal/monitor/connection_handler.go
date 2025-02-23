package monitor

import (
	"net"
	"slices"
	"sync"
)

type ConnectionHandler struct {
	sync.RWMutex
	Clients []*net.TCPConn
}

var connHandler *ConnectionHandler
var connHandlerOnce sync.Once

func GetConnectionHandler() *ConnectionHandler {
	connHandlerOnce.Do(func() {
		connHandler = &ConnectionHandler{
			Clients: make([]*net.TCPConn, 0),
		}
	})

	return connHandler
}

func (c *ConnectionHandler) AddClient(client *net.TCPConn) {
	c.Lock()
	defer c.Unlock()
	c.Clients = append(c.Clients, client)
}

func (c *ConnectionHandler) RemoveClient(client *net.TCPConn) {
	c.Lock()
	defer c.Unlock()
	c.Clients = slices.DeleteFunc(c.Clients, func(conn *net.TCPConn) bool {
		return conn == client
	})
}

func (c *ConnectionHandler) RemoveClientsUnsafe(clients []*net.TCPConn) {
	// Build a map for fast lookups.
	clientsToRemove := make(map[*net.TCPConn]struct{}, len(clients))
	for _, conn := range clients {
		clientsToRemove[conn] = struct{}{}
	}
	// Filter out the clients to be removed
	c.Clients = slices.DeleteFunc(c.Clients, func(conn *net.TCPConn) bool {
		_, exists := clientsToRemove[conn]
		return exists
	})
}

func (c *ConnectionHandler) RemoveClients(clients []*net.TCPConn) {
	// Build a map for fast lookups.
	clientsToRemove := make(map[*net.TCPConn]struct{}, len(clients))
	for _, conn := range clients {
		clientsToRemove[conn] = struct{}{}
	}

	c.Lock()
	defer c.Unlock()
	// Filter out the clients to be removed
	c.Clients = slices.DeleteFunc(c.Clients, func(conn *net.TCPConn) bool {
		_, exists := clientsToRemove[conn]
		return exists
	})
}

func (c *ConnectionHandler) CloseAndRemoveClients() {
	for _, client := range c.Clients {
		c.Lock()
		_ = client.Close()
		c.Unlock()
		c.RemoveClient(client)
	}
}

func (c *ConnectionHandler) SendMessage(s string) {
	c.RLock()

	var clientsToRemove []*net.TCPConn
	mu := sync.Mutex{}

	wg := sync.WaitGroup{}
	wg.Add(len(c.Clients))

	for _, client := range c.Clients {
		go func(conn *net.TCPConn) {
			defer wg.Done()
			if _, err := conn.Write([]byte(s)); err != nil {
				mu.Lock()
				clientsToRemove = append(clientsToRemove, conn)
				mu.Unlock()
			}
		}(client)
	}
	wg.Wait()
	c.RemoveClientsUnsafe(clientsToRemove)
	c.RUnlock()

}
