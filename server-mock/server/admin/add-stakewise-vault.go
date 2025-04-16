package admin

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Add a StakeWise vault to the server
func (s *AdminServer) addStakeWiseVault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	deploymentID := query.Get("deployment")
	if deploymentID == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing deployment query parameter"))
		return
	}
	name := query.Get("name")
	if name == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing name query parameter"))
		return
	}
	addressString := query.Get("address")
	if addressString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing address query parameter"))
		return
	}
	address := ethcommon.HexToAddress(addressString)

	// Create a new vault
	db := s.manager.GetDatabase()
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	deployment.AddVault(name, address)
	s.logger.Info("Added new stakewise vault",
		"deployment", deploymentID,
		"address", address.Hex(),
	)
	common.HandleSuccess(w, s.logger, "")
}
