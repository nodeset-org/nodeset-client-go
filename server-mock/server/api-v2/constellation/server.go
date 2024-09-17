package v2server_constellation

import (
	"log/slog"

	"github.com/gorilla/mux"
	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"

	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v2 server mock for constellation module routes
type V2ConstellationServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v2 constellation server mock
func NewV2ConstellationServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *V2ConstellationServer {
	return &V2ConstellationServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V2ConstellationServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V2ConstellationServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V2ConstellationServer) RegisterRoutes(versionRouter *mux.Router) {
	constellationPrefix := "/" + v2constellation.ConstellationPrefix + "{deployment}/"
	versionRouter.HandleFunc(constellationPrefix+v2constellation.WhitelistPath, s.handleWhitelist)
	versionRouter.HandleFunc(constellationPrefix+v2constellation.MinipoolDepositSignaturePath, s.minipoolDepositSignature)
	versionRouter.HandleFunc(constellationPrefix+v2constellation.ValidatorsPath, s.handleValidators)
}
