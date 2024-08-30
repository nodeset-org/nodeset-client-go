package v0server

import (
	"net/http"

	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/validators
func (s *V0Server) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPatch:
		s.patchValidators(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/validators
func (s *V0Server) getValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	args, _ := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Get the registered validators
	db := s.manager.GetDatabase()
	network := args.Get("network")
	deployment := db.StakeWise.GetDeployment(network)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, network)
		return
	}
	validatorStatuses := []stakewise.ValidatorStatus{}
	validatorsForDeployment := node.GetAllStakeWiseValidators(deployment)

	// Iterate the validators
	for _, validatorsForVault := range validatorsForDeployment {
		for _, validator := range validatorsForVault {
			validatorStatuses = append(validatorStatuses, stakewise.ValidatorStatus{
				Pubkey:              validator.Pubkey,
				Status:              validator.GetStatus(),
				ExitMessageUploaded: validator.ExitMessageUploaded,
			})
		}
	}

	// Write the response
	data := stakewise.ValidatorsData{
		Validators: validatorStatuses,
	}
	common.HandleSuccess(w, s.logger, data)
}

// PATCH api/validators
func (s *V0Server) patchValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var exitData []clientcommon.ExitData
	args, _ := common.ProcessApiRequest(s, w, r, &exitData)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Just get the first vault for the deployment
	database := s.manager.GetDatabase()
	network := args.Get("network")
	deployment := database.StakeWise.GetDeployment(network)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, network)
		return
	}
	vaults := deployment.GetStakeWiseVaults()
	if len(vaults) == 0 {
		common.HandleInvalidDeployment(w, s.logger, network)
		return
	}
	var vault *db.StakeWiseVault
	for _, v := range vaults {
		vault = v
		break
	}

	// Handle the upload
	err := vault.HandleSignedExitUpload(node, exitData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
