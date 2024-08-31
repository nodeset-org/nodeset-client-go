package v2server_core

import (
	"log/slog"

	"github.com/gorilla/mux"
	v2core "github.com/nodeset-org/nodeset-client-go/api-v2/core"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v2 server mock for core routes
type V2CoreServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v2 core server mock
func NewV2CoreServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *V2CoreServer {
	return &V2CoreServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V2CoreServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V2CoreServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V2CoreServer) RegisterRoutes(versionRouter *mux.Router) {
	corePrefix := "/" + v2core.CorePrefix
	versionRouter.HandleFunc(corePrefix+core.LoginPath, s.login)
	versionRouter.HandleFunc(corePrefix+core.NoncePath, s.getNonce)
	versionRouter.HandleFunc(corePrefix+core.NodeAddressPath, s.nodeAddress)
}
