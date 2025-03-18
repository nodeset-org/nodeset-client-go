package v3server

import (
	"log/slog"

	"github.com/gorilla/mux"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
	v2server_constellation "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v2/constellation"
	v2server_core "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v2/core"
	v2server_stakewise "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v2/stakewise"
)

// API v2 server mock
type V2Server struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager

	// Sub-servers
	Core          *v2server_core.V2CoreServer
	StakeWise     *v2server_stakewise.V2StakeWiseServer
	Constellation *v2server_constellation.V2ConstellationServer
}

// Creates a new API v2 server mock
func NewV2Server(logger *slog.Logger, manager *manager.NodeSetMockManager) *V2Server {
	server := &V2Server{
		logger:        logger,
		manager:       manager,
		Core:          v2server_core.NewV2CoreServer(logger, manager),
		StakeWise:     v2server_stakewise.NewV2StakeWiseServer(logger, manager),
		Constellation: v2server_constellation.NewV2ConstellationServer(logger, manager),
	}
	return server
}

// Gets the logger
func (s *V2Server) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V2Server) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V2Server) RegisterRoutes(apiRouter *mux.Router) {
	versionPrefix := "/" + apiv2.ApiVersion
	versionRouter := apiRouter.PathPrefix(versionPrefix).Subrouter()
	s.Core.RegisterRoutes(versionRouter)
	s.StakeWise.RegisterRoutes(versionRouter)
	s.Constellation.RegisterRoutes(versionRouter)
}
