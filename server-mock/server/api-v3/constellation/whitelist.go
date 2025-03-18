package v3server_constellation

import (
	"fmt"
	"net/http"

	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// Handler for api/v2/modules/constellation/{deployment}/whitelist
func (s *V2ConstellationServer) handleWhitelist(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getWhitelist(w, r)
	case http.MethodPost:
		s.postWhitelist(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

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

	// Get the registered address
	email := node.GetUser().Email
	registeredAddress := deployment.GetWhitelistedAddressForUser(email)

	// Write the data
	data := v2constellation.Whitelist_GetData{
		Whitelisted: registeredAddress != nil,
	}
	if data.Whitelisted {
		data.Address = *registeredAddress
	}
	common.HandleSuccess(w, s.logger, data)
}

// POST api/v2/modules/constellation/{deployment}/whitelist
func (s *V2ConstellationServer) postWhitelist(w http.ResponseWriter, r *http.Request) {
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
	signature, err := deployment.GetWhitelistSignature(node.Address)
	if err != nil {
		common.HandleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}
	s.logger.Info("Created Constellation whitelist signature")

	// Write the data
	data := v2constellation.Whitelist_PostData{
		Signature: utils.EncodeHexWithPrefix(signature),
	}
	common.HandleSuccess(w, s.logger, data)
}
