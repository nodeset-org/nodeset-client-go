package server

import (
	"net/http"

	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
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
	validatorStatuses := []apiv1.ValidatorStatus{}
	validatorsForNetwork := node.Validators[network]

	// Iterate the validators
	for _, validator := range validatorsForNetwork {
		pubkey := validator.Pubkey
		status := s.manager.GetValidatorStatus(network, pubkey)
		validatorStatuses = append(validatorStatuses, apiv1.ValidatorStatus{
			Pubkey:              pubkey,
			Status:              status,
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := apiv1.ValidatorsData{
		Validators: validatorStatuses,
	}
	handleSuccess(w, s.logger, data)
}
