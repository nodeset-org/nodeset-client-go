package v3server_core

import (
	"log/slog"

	"github.com/gorilla/mux"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v2 server mock for core routes
type V3CoreServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v2 core server mock
func NewV3CoreServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *V3CoreServer {
	return &V3CoreServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V3CoreServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V3CoreServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V3CoreServer) RegisterRoutes(versionRouter *mux.Router) {
	corePrefix := "/" + v3core.CorePrefix
	versionRouter.HandleFunc(corePrefix+core.LoginPath, s.login)
	versionRouter.HandleFunc(corePrefix+core.NoncePath, s.getNonce)
	versionRouter.HandleFunc(corePrefix+core.NodeAddressPath, s.nodeAddress)
}
