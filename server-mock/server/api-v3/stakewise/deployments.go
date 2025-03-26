package v3server_stakewise

import (
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
	servermockcommon "github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/stakewise/deployments
func (s *V3StakeWiseServer) handleDeployments(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getDeployments(w, r)
	default:
		servermockcommon.HandleInvalidMethod(w, s.logger)
	}
}

func (s *V3StakeWiseServer) getDeployments(w http.ResponseWriter, r *http.Request) {
	// Get the database
	db := s.manager.GetDatabase()

	// Collect deployments
	deployments := []common.Deployment{}
	for _, deployment := range db.StakeWise.Deployments {
		deployments = append(deployments, common.Deployment{
			ChainID: deployment.ChainID.String(),
			Name:    deployment.ID,
		})
	}

	// Return the deployments
	resp := common.DeploymentsData{
		Deployments: deployments,
	}
	servermockcommon.HandleSuccess(w, s.logger, resp)
}
