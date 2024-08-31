package v2server_core_test

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server"
)

const (
	// The timeout for all requests
	timeout time.Duration = 5 * time.Second
)

// Various singleton variables used for testing
var (
	logger *slog.Logger                = slog.Default()
	s      *server.NodeSetMockServer   = nil
	mgr    *manager.NodeSetMockManager = nil
	wg     *sync.WaitGroup             = nil
	port   uint16                      = 0
)

// Initialize a common server used by all tests
func TestMain(m *testing.M) {
	// Create the server
	var err error
	s, err = server.NewNodeSetMockServer(logger, "localhost", 0)
	if err != nil {
		fail("error creating server: %v", err)
	}
	logger.Info("Created server")

	// Start it
	wg = &sync.WaitGroup{}
	err = s.Start(wg)
	if err != nil {
		fail("error starting server: %v", err)
	}
	port = s.GetPort()
	logger.Info(fmt.Sprintf("Started server on port %d", port))
	mgr = s.GetManager()

	// Run tests
	code := m.Run()

	// Revert to the baseline after testing is done
	cleanup()

	// Done
	os.Exit(code)
}

func fail(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	logger.Error(msg)
	cleanup()
	os.Exit(1)
}

func cleanup() {
	if s != nil {
		_ = s.Stop()
		wg.Wait()
		logger.Info("Stopped server")
	}
}
