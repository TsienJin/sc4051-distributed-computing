package logging

import (
	"fmt"
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
			continue
		}
		slog.Info(fmt.Sprintf("[CLIENT:%s] %s", c.RemoteAddr().String(), string(buf[:n])))
	}
}
