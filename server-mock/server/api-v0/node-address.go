package v0server

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// POST api/node-address
func (s *V0Server) nodeAddress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the requesting node
	var request core.NodeAddressRequest
	_, _ = common.ProcessApiRequest(s, w, r, &request)

	// Get the node
	db := s.manager.GetDatabase()
	address := ethcommon.HexToAddress(request.NodeAddress)
	node, isRegistered := db.Core.GetNode(address)
	if node == nil {
		common.HandleNodeNotInWhitelist(w, s.logger, address)
		return
	}
	if isRegistered {
		common.HandleAlreadyRegisteredNode(w, s.logger, address)
		return
	}

	// Register the node
	sig, err := utils.DecodeHex(request.Signature)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid signature"))
		return
	}
	err = node.Register(sig, apiv0.NodeAddressMessageFormat)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Registered new node account", "email", request.Email, "address", address.Hex())
	common.HandleSuccess(w, s.logger, "")
}
