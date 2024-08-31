package admin

import (
	"log/slog"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// Admin routes for the server mock
type AdminServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v0 server mock
func NewAdminServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *AdminServer {
	return &AdminServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *AdminServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *AdminServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *AdminServer) RegisterRoutes(adminRouter *mux.Router) {
	adminRouter.HandleFunc("/"+api.AdminAddConstellationDeploymentPath, s.addConstellationDeployment)
	adminRouter.HandleFunc("/"+api.AdminAddStakeWiseDeploymentPath, s.addStakeWiseDeployment)
	adminRouter.HandleFunc("/"+api.AdminSnapshotPath, s.snapshot)
	adminRouter.HandleFunc("/"+api.AdminRevertPath, s.revert)
	adminRouter.HandleFunc("/"+api.AdminCycleSetPath, s.cycleSet)
	adminRouter.HandleFunc("/"+api.AdminAddUserPath, s.addUser)
	adminRouter.HandleFunc("/"+api.AdminWhitelistNodePath, s.whitelistNode)
	adminRouter.HandleFunc("/"+api.AdminAddVaultPath, s.addStakeWiseVault)
	adminRouter.HandleFunc("/"+api.AdminSetConstellationPrivateKeyPath, s.setConstellationAdminPrivateKey)
	adminRouter.HandleFunc("/"+api.AdminIncrementWhitelistNoncePath, s.incrementWhitelistNonce)
	adminRouter.HandleFunc("/"+api.AdminIncrementSuperNodeNoncePath, s.incrementSuperNodeNonce)
}
