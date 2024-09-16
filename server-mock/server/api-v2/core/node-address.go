package v2server_core

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	v2core "github.com/nodeset-org/nodeset-client-go/api-v2/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// POST api/v2/core/node-address
func (s *V2CoreServer) nodeAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	var request v2core.NodeAddressRequest
	_, _ = common.ProcessApiRequest(s, w, r, &request)
	address := ethcommon.HexToAddress(request.NodeAddress)
	sig, err := utils.DecodeHex(request.Signature)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid signature"))
		return
	}

	// Get the requesting node
	database := s.manager.GetDatabase()
	node, isRegistered := database.Core.GetNode(address)
	if node == nil {
		common.HandleNodeNotInWhitelist(w, s.logger, address)
		return
	}
	if isRegistered {
		common.HandleAlreadyRegisteredNode(w, s.logger, address)
		return
	}

	// Register the node
	err = node.Register(sig, v2core.NodeAddressMessageFormat)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Registered new node account",
		"email", request.Email,
		"address", address.Hex(),
	)
	common.HandleSuccess(w, s.logger, "")
}
