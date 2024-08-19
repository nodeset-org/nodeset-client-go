package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Get the login request
	var request core.LoginRequest
	_ = s.processApiRequest(w, r, &request)
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}

	// Log it in
	address := common.HexToAddress(request.Address)
	signature, err := utils.DecodeHex(request.Signature)
	if err != nil {
		handleInputError(w, s.logger, fmt.Errorf("invalid signature"))
		return
	}
	err = s.manager.Login(request.Nonce, address, signature)
	if err != nil {
		if errors.Is(err, db.ErrUnregisteredNode) {
			handleUnregisteredNode(w, s.logger, address)
			return
		}
		handleServerError(w, s.logger, err)
		return
	}

	// Respond
	data := core.LoginData{
		Token: session.Token,
	}
	handleSuccess(w, s.logger, data)
	s.logger.Info("Logged into session", "nonce", request.Nonce, "address", address.Hex())
}
