package v3server_stakewise

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	stakewiseapi "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPost:
		s.postValidators(w, r)

	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// POST api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) postValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body stakewiseapi.Validators_PostBody

	_, pathArgs := common.ProcessApiRequest(s, w, r, &body)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, deploymentID, vaultAddress)
		return
	}

	numToRegister := len(body.Validators)
	available := int(deployment.MaxValidators) - int(deployment.ActiveValidators)
	if numToRegister > available {
		common.HandleServerError(w, s.logger, fmt.Errorf("not enough available slots: requested %d, available %d", numToRegister, available))
		return
	}
	deployment.ActiveValidators += uint(numToRegister)

	resp := stakewiseapi.PostValidatorData{
		Signature: "TODO",
	}
	common.HandleSuccess(w, s.logger, resp)

}

// GET api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) getValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	_, pathArgs := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, deploymentID, vaultAddress)
		return
	}

	// Find the validator
	validatorStatuses := []stakewise.ValidatorStatus{}
	validators := vault.GetStakeWiseValidatorsForNode(node)
	for _, validator := range validators {
		validatorStatuses = append(validatorStatuses, stakewise.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			Status:              validator.GetStatus(),
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := stakewise.ValidatorsData{
		Validators: validatorStatuses,
	}
	common.HandleSuccess(w, s.logger, data)
}
