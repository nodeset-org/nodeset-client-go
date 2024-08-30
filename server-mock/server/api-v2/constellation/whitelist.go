package v2server_constellation

import (
	"fmt"
	"net/http"

	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// GET api/v2/modules/constellation/{deployment}/whitelist
func (s *V2ConstellationServer) getWhitelist(w http.ResponseWriter, r *http.Request) {
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
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.Constellation.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}

	// Get the signature
	signature, err := deployment.GetConstellationWhitelistSignature(node.Address)
	if err != nil {
		common.HandleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}
	s.logger.Info("Created Constellation whitelist signature")

	// Write the data
	data := v2constellation.WhitelistData{
		Signature: utils.EncodeHexWithPrefix(signature),
	}
	common.HandleSuccess(w, s.logger, data)
}
