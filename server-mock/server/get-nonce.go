package server

import (
	"net/http"

	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
)

func (s *NodeSetMockServer) getNonce(w http.ResponseWriter, r *http.Request) {
	// Create a new session
	session := s.manager.CreateSession()

	// Write the response
	data := apiv1.NonceData{
		Nonce: session.Nonce,
		Token: session.Token,
	}
	handleSuccess(w, s.logger, data)
	s.logger.Info("Created session", "nonce", session.Nonce)
}
