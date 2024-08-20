package v2server_constellation

import (
	"fmt"
	"net/http"

	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
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
	deployment := pathArgs["deployment"]
	if deployment != test.Network { // TEMP
		common.HandleInvalidDeployment(w, s.logger, deployment)
		return
	}

	// Get the signature
	time, signature, err := s.manager.GetConstellationWhitelistSignatureAndTime(node.Address, test.ChainIDBig, test.WhitelistAddress)
	if err != nil {
		common.HandleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}
	s.logger.Info("Created Constellation whitelist signature")

	// Write the data
	data := v2constellation.WhitelistData{
		Signature: utils.EncodeHexWithPrefix(signature),
		Time:      time.Unix(),
	}
	common.HandleSuccess(w, s.logger, data)
}
