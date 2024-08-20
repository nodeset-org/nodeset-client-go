package v2server_constellation

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// POST api/v2/modules/constellation/{deployment}/minipool/deposit-signature
func (s *V2ConstellationServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Parse the request
	request := v2constellation.MinipoolDepositSignatureRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}
	_, pathArgs := common.ProcessApiRequest(s, w, r, nil)

	// Input validation
	deployment := pathArgs["deployment"]
	if deployment != test.Network { // TEMP
		common.HandleInvalidDeployment(w, s.logger, deployment)
		return
	}

	// Prep the args
	salt, success := new(big.Int).SetString(request.Salt, 10)
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
	time, signature, err := s.manager.GetConstellationDepositSignatureAndTime(node.Address, request.MinipoolAddress, salt, test.SuperNodeAddress, test.ChainIDBig)
	if err != nil {
		common.HandleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}

	// Write the data
	data := v2constellation.MinipoolDepositSignatureData{
		Signature: utils.EncodeHexWithPrefix(signature),
		Time:      time.Unix(),
	}
	s.logger.Info("Fetched minipool deposit signature")
	common.HandleSuccess(w, s.logger, data)
}