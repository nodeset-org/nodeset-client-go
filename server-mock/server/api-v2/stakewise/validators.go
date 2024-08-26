package v2server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	v2stakewise "github.com/nodeset-org/nodeset-client-go/api-v2/stakewise"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
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

	// Get the registered validators
	deploymentID := pathArgs["deployment"]
	deployment := s.manager.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	validatorStatuses := []stakewise.ValidatorStatus{}
	validatorsForDeployment := node.Validators[deployment.DeploymentID]

	// Iterate the validators
	for _, validator := range validatorsForDeployment {
		pubkey := validator.Pubkey
		status := s.manager.GetValidatorStatus(deployment.DeploymentID, pubkey)
		validatorStatuses = append(validatorStatuses, stakewise.ValidatorStatus{
			Pubkey:              pubkey,
			Status:              status,
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := stakewise.ValidatorsData{
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
	deploymentID := pathArgs["deployment"]
	deployment := s.manager.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vault := pathArgs["vault"]
	vaultAddress := ethcommon.HexToAddress(vault)

	// Handle the upload
	castedExitData := make([]clientcommon.ExitData, len(body.ExitData))
	for i, data := range body.ExitData {
		castedExitData[i] = clientcommon.ExitData{
			Pubkey: data.Pubkey,
			ExitMessage: clientcommon.ExitMessage{
				Message:   clientcommon.ExitMessageDetails(data.ExitMessage.Message),
				Signature: data.ExitMessage.Signature,
			},
		}
	}
	err := s.manager.HandleSignedExitUpload(node.Address, deploymentID, vaultAddress, castedExitData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
