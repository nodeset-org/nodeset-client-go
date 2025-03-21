package v3server_stakewise

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
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
	var body v3stakewise.Validators_PostBody

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

	// Must add validator to struct + exit message

	// TODO: Confirm with JC
	// NICE TO HAVE: https://github.com/stakewise/v3-core/blob/main/contracts/validators/ValidatorsChecker.sol#L187
	hash := crypto.Keccak256Hash([]byte(fmt.Sprintf("%s:%d", deployment.ID, deployment.ActiveValidators)))
	resp := v3stakewise.PostValidatorData{
		Signature: hash.Hex(), //solidity code for stakewise
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
	validatorStatuses := []v3stakewise.ValidatorStatus{}
	validators := vault.GetStakeWiseValidatorsForNode(node)
	for _, validator := range validators {
		validatorStatuses = append(validatorStatuses, v3stakewise.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := v3stakewise.ValidatorsData{
		Validators: validatorStatuses,
	}
	common.HandleSuccess(w, s.logger, data)
}
