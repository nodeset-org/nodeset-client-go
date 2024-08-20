package v0server

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v0 server mock
type V0Server struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v0 server mock
func NewV0Server(logger *slog.Logger, manager *manager.NodeSetMockManager) *V0Server {
	return &V0Server{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V0Server) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V0Server) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V0Server) RegisterRoutes(apiRouter *mux.Router) {
	// Core
	apiRouter.HandleFunc("/"+core.LoginPath, s.login)
	apiRouter.HandleFunc("/"+core.NoncePath, s.getNonce)
	apiRouter.HandleFunc("/"+core.NodeAddressPath, s.nodeAddress)

	// StakeWise
	apiRouter.HandleFunc("/"+stakewise.DepositDataMetaPath, s.depositDataMeta)
	apiRouter.HandleFunc("/"+stakewise.DepositDataPath, s.handleDepositData)
	apiRouter.HandleFunc("/"+stakewise.ValidatorsPath, s.handleValidators)
}
