package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"server/internal/client/tui"
)

var (
	host string
	port int
)

var rootCmd = &cobra.Command{
	Use:   "client",
	Short: "TUI client written in go for SC4051",
	Run: func(cmd *cobra.Command, args []string) {
		// Validate the host
		if host == "" {
			log.Fatalf("Invalid host: %s", host)
		}

		// Validate port range
		if port < 1 || port > 65535 {
			log.Fatalf("Invalid port number: %d. Must be between 1-65535", port)
		}

		fmt.Printf("Connecting to %s on port %d...\n", host, port)
	},
}

func init() {
	// Define flags
	rootCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "Host to connect to")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "Port number")

	// Ensure both flags are required (alternative: provide default values)
	err := rootCmd.MarkPersistentFlagRequired("host")
	if err != nil {
		log.Fatal(err)
		return
	}
	err = rootCmd.MarkPersistentFlagRequired("port")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func main() {

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}

	tui.StartClient(host, port)

}
