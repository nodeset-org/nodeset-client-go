package v0server

import (
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// GET api/nonce
func (s *V0Server) getNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Create a new session
	db := s.manager.GetDatabase()
	session := db.Core.CreateSession()

	// Write the response
	data := core.NonceData{
		Nonce: session.Nonce,
		Token: session.Token,
	}
	common.HandleSuccess(w, s.logger, data)
	s.logger.Info("Created session", "nonce", session.Nonce)
}
