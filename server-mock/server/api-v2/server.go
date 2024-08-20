package v2server

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v2 server mock
type V2Server struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v2 server mock
func NewV2Server(logger *slog.Logger, manager *manager.NodeSetMockManager) *V2Server {
	return &V2Server{
		logger:  logger,
		manager: manager,
	}
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
	// Core
	apiRouter.HandleFunc("/"+core.LoginPath, s.login)
	apiRouter.HandleFunc("/"+core.NoncePath, s.getNonce)
	apiRouter.HandleFunc("/"+core.NodeAddressPath, s.nodeAddress)

	// StakeWise
	apiRouter.HandleFunc("/"+stakewise.DepositDataMetaPath, s.depositDataMeta)
	apiRouter.HandleFunc("/"+stakewise.DepositDataPath, s.handleDepositData)
	apiRouter.HandleFunc("/"+stakewise.ValidatorsPath, s.handleValidators)
}
