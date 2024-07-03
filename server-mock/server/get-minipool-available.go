package server

import (
	"net/http"

	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
)

func (s *NodeSetMockServer) getMinipoolAvailable(w http.ResponseWriter, r *http.Request) {
	data := apiv2.MinipoolAvailableData{}
	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched minipool available")

}
