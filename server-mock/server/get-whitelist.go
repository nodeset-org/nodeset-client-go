package server

import (
	"net/http"

	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
)

func (s *NodeSetMockServer) getWhitelist(w http.ResponseWriter, r *http.Request) {
	data := apiv2.WhitelistData{}
	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched Constellation whitelist")

}
