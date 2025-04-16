package v3server

import (
	"log/slog"

	"github.com/gorilla/mux"
	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
	v3server_constellation "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v3/constellation"
	v3server_core "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v3/core"
	v3server_stakewise "github.com/nodeset-org/nodeset-client-go/server-mock/server/api-v3/stakewise"
)

// API v3 server mock
type V3Server struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager

	// Sub-servers
	Core          *v3server_core.V3CoreServer
	StakeWise     *v3server_stakewise.V3StakeWiseServer
	Constellation *v3server_constellation.V3ConstellationServer
}

// Creates a new API v3 server mock
func NewV3Server(logger *slog.Logger, manager *manager.NodeSetMockManager) *V3Server {
	server := &V3Server{
		logger:        logger,
		manager:       manager,
		Core:          v3server_core.NewV3CoreServer(logger, manager),
		StakeWise:     v3server_stakewise.NewV3StakeWiseServer(logger, manager),
		Constellation: v3server_constellation.NewV3ConstellationServer(logger, manager),
	}
	return server
}

// Gets the logger
func (s *V3Server) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V3Server) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V3Server) RegisterRoutes(apiRouter *mux.Router) {
	versionPrefix := "/" + apiv3.ApiVersion
	versionRouter := apiRouter.PathPrefix(versionPrefix).Subrouter()
	s.Core.RegisterRoutes(versionRouter)
	s.StakeWise.RegisterRoutes(versionRouter)
	s.Constellation.RegisterRoutes(versionRouter)
}
