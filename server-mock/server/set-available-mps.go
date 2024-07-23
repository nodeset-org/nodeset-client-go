package server

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
)

func (s *NodeSetMockServer) setAvailableConstellationMinipoolCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Get the login request
	var request api.AdminSetAvailableConstellationMinipoolCountRequest
	_ = s.processApiRequest(w, r, &request)
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}

	// Set the count
	err := s.manager.SetAvailableConstellationMinipoolCount(request.UserEmail, request.Count)
	if err != nil {
		handleServerError(w, s.logger, fmt.Errorf("error setting available minipool count: %w", err))
		return
	}
	s.logger.Info("Set available minipool count", "user", request.UserEmail, "count", request.Count)
	handleSuccess(w, s.logger, "")
}
