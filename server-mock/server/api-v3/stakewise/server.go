package v3server_stakewise

import (
	"log/slog"

	"github.com/gorilla/mux"
	v2stakewise "github.com/nodeset-org/nodeset-client-go/api-v2/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"

	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v2 server mock for stakewise module routes
type V2StakeWiseServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v2 stakewise server mock
func NewV2StakeWiseServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *V2StakeWiseServer {
	return &V2StakeWiseServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V2StakeWiseServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V2StakeWiseServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V2StakeWiseServer) RegisterRoutes(versionRouter *mux.Router) {
	stakeWisePrefix := "/" + v2stakewise.StakeWisePrefix + "{deployment}/{vault}/"
	versionRouter.HandleFunc(stakeWisePrefix+stakewise.DepositDataMetaPath, s.depositDataMeta)
	versionRouter.HandleFunc(stakeWisePrefix+stakewise.DepositDataPath, s.handleDepositData)
	versionRouter.HandleFunc(stakeWisePrefix+stakewise.ValidatorsPath, s.handleValidators)
}
