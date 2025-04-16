package v3server_constellation

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/goccy/go-json"

	v3constellation "github.com/nodeset-org/nodeset-client-go/api-v3/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// POST api/v3/modules/constellation/{deployment}/minipool/deposit-signature
func (s *V3ConstellationServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Parse the request
	request := v3constellation.MinipoolDepositSignatureRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	_, pathArgs := common.ProcessApiRequest(s, w, r, nil)

	// Input validation
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.Constellation.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}

	// Prep the args
	salt, success := new(big.Int).SetString(request.Salt, 16)
	if !success {
		common.HandleInputError(w, s.logger, fmt.Errorf("error decoding salt"))
		return
	}

	// Get the requesting node
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Get the signature
	signature, err := deployment.GetMinipoolDepositSignature(node.Address, request.MinipoolAddress, salt)
	if err != nil {
		common.HandleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}

	// Write the data
	data := v3constellation.MinipoolDepositSignatureData{
		Signature: utils.EncodeHexWithPrefix(signature),
	}
	s.logger.Info("Fetched minipool deposit signature")
	common.HandleSuccess(w, s.logger, data)
}
