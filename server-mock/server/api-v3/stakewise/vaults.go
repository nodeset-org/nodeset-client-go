package v3server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"

	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"

	servermockcommon "github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/stakewise/{deployment}/vaults
func (s *V3StakeWiseServer) handleVaults(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getVaults(w, r)
	default:
		servermockcommon.HandleInvalidMethod(w, s.logger)
	}
}

func (s *V3StakeWiseServer) getVaults(w http.ResponseWriter, r *http.Request) {
	// Parse deployment ID from URL
	_, pathArgs := servermockcommon.ProcessApiRequest(s, w, r, nil)
	deploymentID := pathArgs["deployment"]

	// Validate deployment
	db := s.manager.GetDatabase()
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		servermockcommon.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}

	vaults := []ethcommon.Address{}
	for addr := range deployment.Vaults {
		vaults = append(vaults, addr)
	}

	// Return as JSON
	servermockcommon.HandleSuccess(w, s.logger, v3stakewise.VaultsData{Vaults: vaults})
}
