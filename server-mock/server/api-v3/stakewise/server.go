package v3server_stakewise

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"

	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
)

// API v3 server mock for stakewise module routes
type V3StakeWiseServer struct {
	logger  *slog.Logger
	manager *manager.NodeSetMockManager
}

// Creates a new API v3 stakewise server mock
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

func (s *V3StakeWiseServer) RegisterRoutes(versionRouter *mux.Router) {
	basePath := "/" + v3stakewise.StakeWisePrefix + "{deployment}/"
	vaultPath := basePath + "{vault}/"

	// Validators Endpoints
	versionRouter.HandleFunc(vaultPath+stakewise.ValidatorsMetaPath, s.handleValidatorsMeta)
	versionRouter.HandleFunc(vaultPath+stakewise.ValidatorsPath, s.handleValidators)

	// Vaults Endpoint
	versionRouter.HandleFunc(basePath+v3stakewise.VaultsPath, s.handleVaults).Methods(http.MethodGet)
}
