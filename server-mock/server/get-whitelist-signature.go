package server

import (
	"fmt"
	"net/http"

	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) getWhitelistSignature(w http.ResponseWriter, r *http.Request) {
	data := apiv2.WhitelistData{}

	// Get the requesting node
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}
	node := s.getNodeForSession(w, session)
	if node == nil {
		return
	}

	// Get the signature
	signature, err := s.manager.GetConstellationWhitelistSignature(node.Address)
	if err != nil {
		handleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	s.logger.Info("Fetched Constellation whitelist")
	handleSuccess(w, s.logger, data)
}
