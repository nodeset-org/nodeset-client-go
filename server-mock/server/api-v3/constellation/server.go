package v3server_constellation

import (
	"log/slog"

	"github.com/gorilla/mux"
	v3constellation "github.com/nodeset-org/nodeset-client-go/api-v3/constellation"

	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v3 server mock for constellation module routes
type V3ConstellationServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v3 constellation server mock
func NewV2ConstellationServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *V3ConstellationServer {
	return &V3ConstellationServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V3ConstellationServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V3ConstellationServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V3ConstellationServer) RegisterRoutes(versionRouter *mux.Router) {
	constellationPrefix := "/" + v3constellation.ConstellationPrefix + "{deployment}/"
	versionRouter.HandleFunc(constellationPrefix+v3constellation.WhitelistPath, s.handleWhitelist)
	versionRouter.HandleFunc(constellationPrefix+v3constellation.MinipoolDepositSignaturePath, s.minipoolDepositSignature)
	versionRouter.HandleFunc(constellationPrefix+v3constellation.ValidatorsPath, s.handleValidators)
}
