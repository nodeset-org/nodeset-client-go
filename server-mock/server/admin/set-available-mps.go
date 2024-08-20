package admin

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Set the available minipool count for a user
func (s *AdminServer) setAvailableConstellationMinipoolCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the login request
	var request api.AdminSetAvailableConstellationMinipoolCountRequest
	_, _ = common.ProcessApiRequest(s, w, r, &request)

	// Set the count
	err := s.manager.SetAvailableConstellationMinipoolCount(request.UserEmail, request.Count)
	if err != nil {
		common.HandleServerError(w, s.logger, fmt.Errorf("error setting available minipool count: %w", err))
		return
	}
	s.logger.Info("Set available minipool count", "user", request.UserEmail, "count", request.Count)
	common.HandleSuccess(w, s.logger, "")
}
