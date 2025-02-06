package logging

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
)

// HandleClientMessages is a simple function to log messages from clients into the same stream.
// This allows for all watches to view client logs (assuming logs are sent to the server as well.
func HandleClientMessages(c *net.TCPConn) {
	buf := make([]byte, 2048)
	for {
		n, err := c.Read(buf)
		if err != nil {
			// Check if the error indicates that the connection is closed.
			if err == io.EOF {
				log.Printf("[CLIENT] Connection from %s closed by the client", c.RemoteAddr().String())
				return // Exit the loop (and function) if the connection is closed.
			}
			// Handle other potential errors.
			log.Printf("[CLIENT] Error reading from %s: %v", c.RemoteAddr().String(), err)
			continue
		}
		slog.Info(fmt.Sprintf("[CLIENT:%s] %s", c.RemoteAddr().String(), string(buf[:n])))
	}
}
