package v2server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	v2stakewise "github.com/nodeset-org/nodeset-client-go/api-v2/stakewise"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v2/modules/stakewise/{deployment}/{vault}/validators
func (s *V2StakeWiseServer) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPatch:
		s.patchValidators(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/v2/modules/stakewise/{deployment}/{vault}/validators
func (s *V2StakeWiseServer) getValidators(w http.ResponseWriter, r *http.Request) {
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
	validatorStatuses := []v2stakewise.ValidatorStatus{}
	validators := vault.GetStakeWiseValidatorsForNode(node)
	for _, validator := range validators {
		validatorStatuses = append(validatorStatuses, v2stakewise.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			Status:              validator.GetStatusV2(),
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := v2stakewise.ValidatorsData{
		Validators: validatorStatuses,
	}
	common.HandleSuccess(w, s.logger, data)
}

// PATCH api/v2/modules/stakewise/{deployment}/{vault}/validators
func (s *V2StakeWiseServer) patchValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body v2stakewise.Validators_PatchBody
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

	// Handle the upload
	castedExitData := make([]clientcommon.EncryptedExitData, len(body.ExitData))
	for i, data := range body.ExitData {
		castedExitData[i] = clientcommon.EncryptedExitData{
			Pubkey:      data.Pubkey,
			ExitMessage: data.ExitMessage,
		}
	}
	err := vault.HandleEncryptedSignedExitUpload(node, castedExitData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
