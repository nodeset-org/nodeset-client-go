package admin

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Add a new user account to the service
func (s *AdminServer) addUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing email query parameter"))
		return
	}

	// Create a new deposit data set
	err := s.manager.AddUser(email)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Added new user", "email", email)
	common.HandleSuccess(w, s.logger, "")
}
