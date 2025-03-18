package v3server_stakewise

import (
	"log/slog"

	"github.com/gorilla/mux"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v2/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"

	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v2 server mock for stakewise module routes
type V3StakeWiseServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v2 stakewise server mock
func NewV3StakeWiseServer(logger *slog.Logger, manager *manager.NodeSetMockManager) *V3StakeWiseServer {
	return &V3StakeWiseServer{
		logger:  logger,
		manager: manager,
	}
}

// Gets the logger
func (s *V3StakeWiseServer) GetLogger() *slog.Logger {
	return s.logger
}

// Gets the manager
func (s *V3StakeWiseServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// Registers the routes for the server
func (s *V3StakeWiseServer) RegisterRoutes(versionRouter *mux.Router) {
	stakeWisePrefix := "/" + v3stakewise.StakeWisePrefix + "{deployment}/{vault}/"
	versionRouter.HandleFunc(stakeWisePrefix+stakewise.DepositDataMetaPath, s.depositDataMeta)
	versionRouter.HandleFunc(stakeWisePrefix+stakewise.DepositDataPath, s.handleDepositData)
	versionRouter.HandleFunc(stakeWisePrefix+stakewise.ValidatorsPath, s.handleValidators)
}
