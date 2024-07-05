package server

import (
	"net/http"

	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
)

func (s *NodeSetMockServer) getMinipoolAvailable(w http.ResponseWriter, r *http.Request) {
	data := apiv2.MinipoolAvailableData{}

	// Get the requesting node
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}
	node := s.getNodeForSession(w, session)
	if node == nil {
		return
	}

	// Get minipool available count from database
	db := db.NewDatabase(s.logger)
	availabilityCount, err := db.GetAvailableConstellationMinipoolCount(node.Address)
	if err != nil {
		s.logger.Error("Error getting available minipool count", "error", err)
	}

	data.Count = availabilityCount
	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched minipool available count")
}
