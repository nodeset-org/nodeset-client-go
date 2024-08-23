package v0server

import (
	"net/http"

	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
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
	network := args.Get("network")
	validatorStatuses := []stakewise.ValidatorStatus{}
	validatorsForNetwork := node.Validators[network]

	// Iterate the validators
	for _, validator := range validatorsForNetwork {
		pubkey := validator.Pubkey
		status := s.manager.GetValidatorStatus(network, pubkey)
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
	network := args.Get("network")
	vaults := s.manager.GetStakeWiseVaults(network)
	if len(vaults) == 0 {
		common.HandleInvalidDeployment(w, s.logger, network)
		return
	}

	// Handle the upload
	err := s.manager.HandleSignedExitUpload(node.Address, network, vaults[0].Address, exitData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
