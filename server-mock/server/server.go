package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/admin"
	v0server "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v0"
	v2server "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v2"
	v3server "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v3"

	"github.com/rocket-pool/node-manager-core/log"
)

type NodeSetMockServer struct {
	logger  *slog.Logger
	ip      string
	port    uint16
	socket  net.Listener
	server  http.Server
	router  *mux.Router
	manager *manager.NodeSetMockManager

	// Route handlers
	adminServer *admin.AdminServer
	apiv0Server *v0server.V0Server
	apiv2Server *v2server.V2Server
	apiv3Server *v3server.V3Server
}

func NewNodeSetMockServer(logger *slog.Logger, ip string, port uint16) (*NodeSetMockServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	server := &NodeSetMockServer{
		logger: logger,
		ip:     ip,
		port:   port,
		router: router,
		server: http.Server{
			Handler: router,
		},
		manager: manager.NewNodeSetMockManager(logger),
	}
	server.adminServer = admin.NewAdminServer(logger, server.manager)
	server.apiv0Server = v0server.NewV0Server(logger, server.manager)
	server.apiv2Server = v2server.NewV2Server(logger, server.manager)

	// Register admin routes
	adminRouter := router.PathPrefix("/admin").Subrouter()
	server.adminServer.RegisterRoutes(adminRouter)

	// Register API routes
	apiRouter := router.PathPrefix("/api").Subrouter()
	server.apiv0Server.RegisterRoutes(apiRouter)
	server.apiv2Server.RegisterRoutes(apiRouter)
	server.apiv3Server.RegisterRoutes(apiRouter)

	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *NodeSetMockServer) Start(wg *sync.WaitGroup) error {
	// Create the socket
	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Get the port if random
	if s.port == 0 {
		s.port = uint16(socket.Addr().(*net.TCPAddr).Port)
	}

	// Start listening
	wg.Add(1)
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error while listening for HTTP requests", log.Err(err))
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (s *NodeSetMockServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

// Get the port the server is listening on
func (s *NodeSetMockServer) GetPort() uint16 {
	return s.port
}

// Get the mock manager for direct access
func (s *NodeSetMockServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}
