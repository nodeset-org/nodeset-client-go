package v2server_core

import (
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// GET api/v2/core/nonce
func (s *V2CoreServer) getNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Create a new session
	session := s.manager.CreateSession()

	// Write the response
	data := core.NonceData{
		Nonce: session.Nonce,
		Token: session.Token,
	}
	common.HandleSuccess(w, s.logger, data)
	s.logger.Info("Created session", "nonce", session.Nonce)
}
