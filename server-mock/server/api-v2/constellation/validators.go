package v2server_constellation

import (
	"net/http"

	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v2/modules/constellation/{deployment}/validators
func (s *V2ConstellationServer) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPatch:
		s.patchValidators(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/v2/modules/constellation/{deployment}/validators
func (s *V2ConstellationServer) getValidators(w http.ResponseWriter, r *http.Request) {
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
	deployment := db.Constellation.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}

	// Get the validators
	validators := deployment.GetValidatorsForNode(node)
	statuses := make([]v2constellation.ValidatorStatus, len(validators))
	for i, validator := range validators {
		statuses[i] = v2constellation.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			RequiresExitMessage: validator.GetExitMessage() == nil,
		}
	}

	// Write the data
	data := v2constellation.ValidatorsData{
		Validators: statuses,
	}
	common.HandleSuccess(w, s.logger, data)
}

// PATCH api/v2/modules/constellation/{deployment}/validators
func (s *V2ConstellationServer) patchValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body v2constellation.Validators_PatchBody
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
	deployment := db.Constellation.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}

	// Get the validators
	validators := deployment.GetValidatorsForNode(node)
	statuses := make([]v2constellation.ValidatorStatus, len(validators))
	for i, validator := range validators {
		statuses[i] = v2constellation.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			RequiresExitMessage: validator.GetExitMessage() == nil,
		}
	}

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
	err := deployment.HandleSignedExitUpload(node, castedExitData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
