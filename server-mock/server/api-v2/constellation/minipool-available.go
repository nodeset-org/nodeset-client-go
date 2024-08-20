package v2server_constellation

import (
	"net/http"

	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// GET api/v2/modules/constellation/{deployment}/minipool/available
func (s *V2ConstellationServer) minipoolAvailable(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	_, pathArgs := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	deployment := pathArgs["deployment"]
	if deployment != test.Network { // TEMP
		common.HandleInvalidDeployment(w, s.logger, deployment)
		return
	}

	// Get minipool available count from database
	availabilityCount, err := s.manager.GetAvailableConstellationMinipoolCount(node.Address)
	if err != nil {
		s.logger.Error("Error getting available minipool count", "error", err)
		return
	}

	// Write the data
	data := v2constellation.MinipoolAvailableData{
		Count: availabilityCount,
	}
	common.HandleSuccess(w, s.logger, data)
	s.logger.Info("Fetched minipool available count")
}
