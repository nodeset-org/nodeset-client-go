package server

import (
	"net/http"

	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
)

func (s *NodeSetMockServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	data := apiv2.MinipoolDepositSignatureData{}
	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched minipool deposit signature")
}
