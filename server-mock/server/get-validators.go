package server

import (
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
)

func (s *NodeSetMockServer) getValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	args := s.processApiRequest(w, r, nil)
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}
	node := s.getNodeForSession(w, session)
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
	handleSuccess(w, s.logger, data)
}
