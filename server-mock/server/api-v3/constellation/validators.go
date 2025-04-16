package v3server_constellation

import (
	"errors"
	"net/http"

	v3constellation "github.com/nodeset-org/nodeset-client-go/api-v3/constellation"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/constellation/{deployment}/validators
func (s *V3ConstellationServer) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPatch:
		s.patchValidators(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/v3/modules/constellation/{deployment}/validators
func (s *V3ConstellationServer) getValidators(w http.ResponseWriter, r *http.Request) {
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
	statuses := make([]v3constellation.ValidatorStatus, len(validators))
	for i, validator := range validators {
		statuses[i] = v3constellation.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			RequiresExitMessage: validator.GetExitMessage() == nil,
		}
	}

	// Write the data
	data := v3constellation.ValidatorsData{
		Validators: statuses,
	}
	common.HandleSuccess(w, s.logger, data)
}

// PATCH api/v3/modules/constellation/{deployment}/validators
func (s *V3ConstellationServer) patchValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body v3constellation.Validators_PatchBody
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
	nsDB := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := nsDB.Constellation.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}

	// Get the validators
	validators := deployment.GetValidatorsForNode(node)
	statuses := make([]v3constellation.ValidatorStatus, len(validators))
	for i, validator := range validators {
		statuses[i] = v3constellation.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			RequiresExitMessage: validator.GetExitMessage() == nil,
		}
	}

	// Handle the upload
	castedExitData := make([]clientcommon.EncryptedExitData, len(body.ExitData))
	for i, data := range body.ExitData {
		castedExitData[i] = clientcommon.EncryptedExitData{
			Pubkey:      data.Pubkey,
			ExitMessage: data.ExitMessage,
		}
	}
	err := deployment.HandleEncryptedSignedExitUpload(node, castedExitData)
	if err != nil {
		if errors.Is(err, db.ErrSignedExitAlreadyUploaded) {
			common.HandleExitAlreadyExists(w, s.logger)
			return
		}
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
