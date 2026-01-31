// test_tcp_discovery.go - Simple TCP discovery test program
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"dev.helix.agent/internal/config"
	"dev.helix.agent/internal/services/discovery"
	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.Int("port", 0, "Port to test TCP discovery")
	flag.Parse()

	if *port == 0 {
		fmt.Println("Error: --port required")
		os.Exit(1)
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs

	discoverer := discovery.NewDiscoverer(logger)

	endpoint := &config.ServiceEndpoint{
		ServiceName:      "test-tcp",
		Host:             "127.0.0.1",
		Port:             fmt.Sprintf("%d", *port),
		DiscoveryEnabled: true,
		DiscoveryMethod:  "tcp",
		DiscoveryTimeout: 2 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	discovered, err := discoverer.Discover(ctx, endpoint)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Discovered: %v\n", discovered)
	if discovered {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
