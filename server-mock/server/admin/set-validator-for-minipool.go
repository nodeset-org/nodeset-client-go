package admin

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Set the validator pubkey for a minipool
func (s *AdminServer) setValidatorForMinipool(w http.ResponseWriter, r *http.Request) {
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
	minipoolString := query.Get("minipool")
	if minipoolString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing minipool query parameter"))
		return
	}
	minipool := ethcommon.HexToAddress(minipoolString)
	validatorString := query.Get("pubkey")
	if validatorString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing pubkey query parameter"))
		return
	}
	pubkey, err := beacon.HexToValidatorPubkey(validatorString)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid pubkey query parameter"))
		return
	}

	// Add a new deployment
	db := s.manager.GetDatabase()
	deployment := db.Constellation.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	deployment.SetValidatorInfoForMinipool(minipool, pubkey)
	s.logger.Info("Validator is now owned by minipool",
		"deployment", deploymentID,
		"minipool", minipool.Hex(),
		"pubkey", pubkey.Hex(),
	)
	common.HandleSuccess(w, s.logger, "")
}
